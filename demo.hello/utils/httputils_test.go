package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

var (
	url = "http://127.0.0.1:8081/ping"
)

func TestBuildURL(t *testing.T) {
	host := "http://127.0.0.1:17891/test"
	query := map[string]string{
		"id":   "1011",
		"name": "foo",
	}
	res, err := BuildURL(host, query)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("encoded url:", res)
}

func TestGetLocalHostIPs(t *testing.T) {
	hosts, err := GetLocalHostIPs()
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("local hosts:", strings.Join(hosts, ", "))
}

func TestGetRemoteHostIPs(t *testing.T) {
	hosts := []string{"https://www.baidu.com/"}
	for _, host := range hosts {
		ips, err := GetRemoteHostIPs(host)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Println("host ips:", strings.Join(ips, ", "))
	}
}

func TestHttpUtilsGet(t *testing.T) {
	utils := NewHTTPUtils(false)
	resp, err := utils.Get(context.TODO(), url, map[string]string{})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("response:", string(resp))
}

func TestHttpUtilsPost(t *testing.T) {
	const (
		url = "http://127.0.0.1:8081/users/new"
		// jsonBody = `{"name": "tester01", "age": 39}`
	)
	jsonMap := map[string]interface{}{
		"name": "tester01",
		"age":  39,
	}
	jsonBody, err := json.Marshal(jsonMap)
	if err != nil {
		t.Fatal(err)
	}

	utils := NewHTTPUtils(false)
	resp, err := utils.Post(context.TODO(), url, map[string]string{}, string(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("response:", string(resp))
}

/*
Number of http connections test.

monitor connections: watch -n 1 "netstat -n | grep 8081"
*/

func TestHttpGet01(t *testing.T) {
	// long client => 1 connection
	utils := NewDefaultHTTPUtils()
	start := time.Now()
	for i := 0; i < 5; i++ {
		if b, err := utils.Get(context.TODO(), url, map[string]string{}); err != nil {
			t.Fatal(err)
		} else {
			fmt.Println("Response:", string(b))
		}
	}
	t.Log("Time:", time.Since(start))
}

func TestHttpGet02(t *testing.T) {
	// short client => 5 connection
	utils := NewHTTPUtils(false)
	start := time.Now()
	for i := 0; i < 5; i++ {
		if b, err := utils.Get(context.TODO(), url, map[string]string{}); err != nil {
			t.Fatal(err)
		} else {
			fmt.Println("Response:", string(b))
		}
	}
	t.Log("Time:", time.Since(start))
}

func TestHttpGet03(t *testing.T) {
	// long client + goroutine => 5 connections
	utils := NewDefaultHTTPUtils()
	wg := sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if b, err := utils.Get(context.TODO(), url, map[string]string{}); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Response:", string(b))
			}
		}()
	}
	wg.Wait()
	t.Log("Time:", time.Since(start))
}

func TestHttpGet04(t *testing.T) {
	// custom long client + goroutine => 5 connections
	utils := NewHTTPUtils(true)
	wg := sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if b, err := utils.Get(context.TODO(), url, map[string]string{}); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Response:", string(b))
			}
		}()
	}
	wg.Wait()
	t.Log("Time:", time.Since(start))
}

/*
Ping a tcp connection.
*/

func TestPingTCP(t *testing.T) {
	ret, err := PingTCP("localhost", "50051")
	fmt.Println("ping result:", ret)
	if err != nil {
		t.Fatal(err)
	}
}
