package utils

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

/*
Http short and long link.
*/

func TestHttpGet01(t *testing.T) {
	// http 短连接
	// monitor connections: watch -n 1 "netstat -n | grep 17891"
	const times = 5
	const url = "http://127.0.0.1:17891/ping"

	client := http.Client{}

	start := time.Now()
	for i := 0; i < times; i++ {
		if err := HTTPGet(client, url); err != nil {
			t.Fatal(err)
		}
	}
	t.Log("Orig Go Net Short Link", time.Since(start))
}

func TestHttpGet02(t *testing.T) {
	// http 短连接 goroutine
	const times = 5
	const url = "http://127.0.0.1:17891/ping"

	var err error
	client := http.Client{}
	wg := sync.WaitGroup{}

	start := time.Now()
	for i := 0; i < times; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err != nil {
				return
			}
			err = HTTPGet(client, url)
		}()
	}
	wg.Wait()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("Orig Go Net Short Link", time.Since(start))
}

func TestHttpGet03(t *testing.T) {
	// http 长连接 goroutine
	const times = 5
	const url = "http://127.0.0.1:17891/ping"

	var err error
	httpTransport := &http.Transport{
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 60 * time.Second,
		}).DialContext,
		MaxIdleConns:          500,              // 最大空闲连接
		IdleConnTimeout:       60 * time.Second, // 空闲连接的超时时间
		ExpectContinueTimeout: 30 * time.Second, // 等待服务第一个响应的超时时间
		MaxIdleConnsPerHost:   100,              // 每个host保持的空闲连接数
	}
	client := http.Client{Transport: httpTransport}
	wg := sync.WaitGroup{}

	start := time.Now()
	for i := 0; i < times; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err != nil {
				return
			}
			err = HTTPGet(client, url)
		}()
	}
	wg.Wait()
	if err != nil {
		t.Fatal(err)
	}
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
