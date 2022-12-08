package http_aop

import (
	"fmt"
	"go1_1711_demo/utils"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"sync"
)

var withMyHttpTransportOnce sync.Once

func WithMyHttpTransport(client *http.Client, transport http.RoundTripper) {
	withMyHttpTransportOnce.Do(func() {
		log.Println("add http interceptor")
		newTransport := NewMyHttpTransport(transport)
		if client == nil {
			http.DefaultTransport = newTransport
		} else {
			client.Transport = newTransport
		}
	})
}

type MyHttpTransport struct {
	proxy http.RoundTripper
}

func NewMyHttpTransport(transport http.RoundTripper) *MyHttpTransport {
	if transport != nil {
		return &MyHttpTransport{
			proxy: transport,
		}
	}

	if defaultTransport, ok := http.DefaultTransport.(*http.Transport); ok {
		return &MyHttpTransport{
			proxy: defaultTransport,
		}
	}
	log.Fatalln("get http default transport error")
	return nil
}

// RoundTrip: impls http.RoundTripper interface.
func (transport *MyHttpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	log.Println("http interceptor: before")
	b, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, fmt.Errorf("dump request error: %v", err)
	}
	log.Printf("dump request: %s", b)

	info, err := getCallerInfo()
	if err != nil {
		log.Println("get caller info failed:", err)
	}
	log.Println("caller info:", info)

	resp, err := transport.proxy.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	log.Println("http interceptor: after")
	b, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, fmt.Errorf("dump response error: %v", err)
	}
	log.Printf("dump response: %s", b)

	return resp, nil
}

// Helper

var (
	ErrNoMatchedCallerInfo = fmt.Errorf("no matched caller info found")

	HttpMethodTags = map[string]struct{}{
		"do":   {},
		"send": {},
		"get":  {},
		"post": {},
	}
)

func getCallerInfo() (string, error) {
	for i := 4; i < 10; i++ {
		info, err := utils.GetCallerInfo(i)
		if err != nil {
			return "", err
		}
		fullFnName := strings.ToLower(info.FnName)
		fnName := fullFnName[strings.LastIndex(fullFnName, ".")+1:]
		if _, ok := HttpMethodTags[fnName]; !ok {
			return fmt.Sprintf("fnname:%s | file:%s | line:%d", info.FnName, info.File, info.LineNo), nil
		}
	}
	return "", ErrNoMatchedCallerInfo
}
