package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"

	pb "demo.grpc/grpc.vs.rest/proto"
	"golang.org/x/net/http2"
	"google.golang.org/grpc"
)

var client http.Client

func init() {
	client = http.Client{}
}

func createTLSConfigWithCustomCert() *tls.Config {
	caCert, err := ioutil.ReadFile("restserver/ca.crt")
	if err != nil {
		log.Fatalf("Reading server certificate: %s", err)
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	return &tls.Config{
		RootCAs: caCertPool,
	}
}

type Request struct {
	Path   string
	Input  *pb.Random
	Output *pb.Random
}

func httpPost(request Request) error {
	b, err := json.Marshal(request.Input)
	if err != nil {
		log.Println("error marshalling input:", err)
		return err
	}

	req, err := http.NewRequest("POST", request.Path, bytes.NewBuffer(b))
	if err != nil {
		log.Println("error creating request:", err)
		return err
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Println("error executing request:", err)
		return err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("error reading response body:", err)
		return err
	}

	if err := json.Unmarshal(bytes, request.Output); err != nil {
		log.Println("error unmarshalling response:", err)
		return err
	}
	return nil
}

/*
http1.1, http2 and grpc api test
*/

func TestHttp11Post(t *testing.T) {
	client.Transport = &http.Transport{
		TLSClientConfig: createTLSConfigWithCustomCert(),
	}

	request := Request{
		Path: "https://zj.ssltest.com:8080",
		Input: &pb.Random{
			RandomString: "random string",
			RandomInt:    1982,
		},
		Output: &pb.Random{},
	}
	if err := httpPost(request); err != nil {
		t.Fatal(err)
	}

	bytes, err := json.Marshal(request.Output)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("http1.1 test response:", string(bytes))
}

func TestHttp11KeepAlivePost(t *testing.T) {
	client.Transport = &http.Transport{
		TLSClientConfig: createTLSConfigWithCustomCert(),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		MaxIdleConns:          500,
		IdleConnTimeout:       time.Duration(60) * time.Second,
		ExpectContinueTimeout: time.Duration(30) * time.Second,
		MaxIdleConnsPerHost:   100,
	}

	request := Request{
		Path: "https://zj.ssltest.com:8080",
		Input: &pb.Random{
			RandomString: "random string",
			RandomInt:    1982,
		},
		Output: &pb.Random{},
	}

	for i := 0; i < 10; i++ {
		request.Input.RandomInt++
		if err := httpPost(request); err != nil {
			t.Fatal(err)
		}

		bytes, err := json.Marshal(request.Output)
		if err != nil {
			t.Fatal(err)
		}
		log.Println("http1.1 keep alive test response:", string(bytes))
	}
}

func TestHttp2Post(t *testing.T) {
	client.Transport = &http2.Transport{
		TLSClientConfig: createTLSConfigWithCustomCert(),
	}

	request := Request{
		Path: "https://zj.ssltest.com:8080",
		Input: &pb.Random{
			RandomString: "random string",
			RandomInt:    2021,
		},
		Output: &pb.Random{},
	}
	if err := httpPost(request); err != nil {
		t.Fatal(err)
	}

	bytes, err := json.Marshal(request.Output)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("http2 test response:", string(bytes))
}

func TestGrpc(t *testing.T) {
	conn, err := grpc.Dial("zj.ssltest.com:9090", grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Grpc dial failed: %v\n", err)
	}
	client := pb.NewRandomServiceClient(conn)

	random := &pb.Random{
		RandomString: "random string",
		RandomInt:    1982,
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	output, err := client.DoSomething(ctx, random)
	if err != nil {
		t.Fatalf("grpc invoke failed: %v\n", err)
	}

	bytes, err := json.Marshal(output)
	if err != nil {
		t.Fatal(err)
	}
	log.Println("grpc test response:", string(bytes))
}

/*
http1.1 and http2 benchmark
*/

func TestChannel(t *testing.T) {
	ch := make(chan int)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(ch chan int, idx int) {
			defer wg.Done()
			count := 0
			for {
				num := <-ch
				if num > 10000 {
					log.Printf("goroutine %d count: %d\n", idx, count)
					return
				}
				count++
			}
		}(ch, i)
	}

	for i := 0; i < 10000; i++ {
		ch <- i
	}
	ch <- 10001
	ch <- 10002
	wg.Wait()

	log.Println("test channel done.")
}

func BenchmarkHTTP11Post(b *testing.B) {
	// benchmark results (with error: 1. EOF; 2. connection reset by peer):
	// BenchmarkHTTP11Get-16    	    9603	    194079 ns/op	   27683 B/op	     315 allocs/op
	client.Transport = &http.Transport{
		TLSClientConfig: createTLSConfigWithCustomCert(),
	}

	requestQueue := make(chan Request)
	defer startWorkers(requestQueue, restWorker)()

	request := Request{
		Path: "https://zj.ssltest.com:8080",
		Input: &pb.Random{
			RandomString: "random string",
		},
		Output: &pb.Random{},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Input.RandomInt = rand.Int31n(100000)
		requestQueue <- request
	}
}

func BenchmarkHTTP11KeepAlivePost(b *testing.B) {
	// benchmark results:
	// BenchmarkHTTP11KeepAlivePost-16    	   31498	     36499 ns/op	    4435 B/op	      56 allocs/op
	client.Transport = &http.Transport{
		TLSClientConfig: createTLSConfigWithCustomCert(),
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		MaxIdleConns:          500,
		IdleConnTimeout:       time.Duration(60) * time.Second,
		ExpectContinueTimeout: time.Duration(30) * time.Second,
		MaxIdleConnsPerHost:   100,
	}

	requestQueue := make(chan Request)
	defer startWorkers(requestQueue, restWorker)()

	request := Request{
		Path: "https://zj.ssltest.com:8080",
		Input: &pb.Random{
			RandomString: "random string",
		},
		Output: &pb.Random{},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Input.RandomInt = rand.Int31n(100000)
		requestQueue <- request
	}
}

func BenchmarkHTTP2Post(b *testing.B) {
	// benchmark results:
	// BenchmarkHTTP2Get-16    	   16494	     66389 ns/op	    8568 B/op	      56 allocs/op
	client.Transport = &http2.Transport{
		TLSClientConfig: createTLSConfigWithCustomCert(),
	}

	requestQueue := make(chan Request)
	defer startWorkers(requestQueue, restWorker)()

	request := Request{
		Path: "https://zj.ssltest.com:8080",
		Input: &pb.Random{
			RandomString: "random string",
		},
		Output: &pb.Random{},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Input.RandomInt = rand.Int31n(100000)
		requestQueue <- request
	}
}

/*
grpc benchmark
*/

func BenchmarkGrpc(b *testing.B) {
	// benchmark results:
	// BenchmarkGrpc-16    	   30862	     38768 ns/op	    4644 B/op	      87 allocs/op
	conn, err := grpc.Dial("zj.ssltest.com:9090", grpc.WithInsecure())
	if err != nil {
		b.Fatalf("Grpc dial failed: %v\n", err)
	}
	client := pb.NewRandomServiceClient(conn)

	requestQueue := make(chan Request)
	defer startWorkers(requestQueue, getGrpcWorker(client))()

	request := Request{
		Path: "https://zj.ssltest.com:9090",
		Input: &pb.Random{
			RandomString: "random string",
		},
		Output: &pb.Random{},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request.Input.RandomInt = rand.Int31n(100000)
		requestQueue <- request
	}
}

/*
workers
*/

const stopRequestPath = "STOP"
const noWorkers = 4

func startWorkers(requestQueue chan Request, worker func(chan Request, *sync.WaitGroup)) func() {
	var wg sync.WaitGroup
	for i := 0; i < noWorkers; i++ {
		wg.Add(1)
		go worker(requestQueue, &wg)
	}

	return func() {
		stopRequest := Request{Path: stopRequestPath}
		for i := 0; i < noWorkers; i++ {
			requestQueue <- stopRequest
		}
		wg.Wait()
	}
}

func restWorker(requestQueue chan Request, wg *sync.WaitGroup) {
	defer wg.Done()
	var count int
	for {
		count++
		request := <-requestQueue
		if request.Path == stopRequestPath {
			log.Printf("worker run count: %d", count)
			return
		}
		httpPost(request)
	}
}

func getGrpcWorker(client pb.RandomServiceClient) func(chan Request, *sync.WaitGroup) {
	return func(requestQueue chan Request, wg *sync.WaitGroup) {
		defer wg.Done()
		var count int
		for {
			request := <-requestQueue
			count++
			if request.Path == stopRequestPath {
				log.Printf("worker run count: %d", count)
				return
			}
			client.DoSomething(context.TODO(), request.Input)
		}
	}
}
