package demos

import (
	"context"
	"fmt"
	"testing"
	"time"

	"demo.hello/utils"
)

func TestStartHTTPServer01(t *testing.T) {
	go func() {
		if err := startHTTPServerV1(); err != nil {
			panic(err)
		}
	}()

	time.Sleep(time.Second)

	client := utils.NewDefaultHTTPUtils()
	host := fmt.Sprintf("http://%s:%d", addr, port)
	for _, path := range [3]string{"/", "/ping", "/ping"} {
		b, err := client.Get(context.Background(), host+path, map[string]string{})
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("[client] resp: %s\n", b)
	}
}

func TestStartHTTPServer02(t *testing.T) {
	go func() {
		if err := startHTTPServerV2(); err != nil {
			panic(err)
		}
	}()

	time.Sleep(time.Second)

	client := utils.NewDefaultHTTPUtils()
	host := fmt.Sprintf("http://%s:%d", addr, port)
	headers := map[string]string{"XTag": "XTest"}
	for _, path := range [2]string{"/", "/ping"} {
		body := fmt.Sprintln("http body test:", path)
		b, err := client.Post(context.Background(), host+path, headers, body)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Printf("[client] resp: %s\n", b)
	}
}
