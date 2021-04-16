package utils

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"testing"
	"time"
)

var (
	url = "http://127.0.0.1:8081/ping"
)

func TestHttpUtilsGet(t *testing.T) {
	utils := NewHTTPUtils(false)
	resp, err := utils.HTTPGet(context.TODO(), url, map[string]string{})
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
	resp, err := utils.HTTPPost(context.TODO(), url, map[string]string{}, string(jsonBody))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("response:", string(resp))
}

/*
Http client test.

monitor connections: watch -n 1 "netstat -n | grep 8081"
*/

func TestHttpGet01(t *testing.T) {
	// default client => 1 connection
	utils := NewHTTPUtils(false)
	start := time.Now()
	for i := 0; i < 5; i++ {
		if b, err := utils.HTTPGet(context.TODO(), url, map[string]string{}); err != nil {
			t.Fatal(err)
		} else {
			fmt.Println("Response:", string(b))
		}
	}
	t.Log("Orig Go Net Short Link", time.Since(start))
}

func TestHttpGet02(t *testing.T) {
	// default client + goroutine => 5 connections
	utils := NewHTTPUtils(false)
	wg := sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if b, err := utils.HTTPGet(context.TODO(), url, map[string]string{}); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Response:", string(b))
			}
		}()
	}
	wg.Wait()
	t.Log("Orig Go Net Short Link", time.Since(start))
}

func TestHttpGet03(t *testing.T) {
	// custom client + goroutine => 5 connections
	utils := NewHTTPUtils(true)
	wg := sync.WaitGroup{}
	start := time.Now()
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if b, err := utils.HTTPGet(context.TODO(), url, map[string]string{}); err != nil {
				fmt.Println(err)
			} else {
				fmt.Println("Response:", string(b))
			}
		}()
	}
	wg.Wait()
	t.Log("Orig GoNet Long Link", time.Since(start))
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
