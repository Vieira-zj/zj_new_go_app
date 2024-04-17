package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// Network

func GetLocalIPAddr() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return "-1:", errors.New("not happen")
}

func GetLocalIPAddrByDial() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return "", err
	}

	if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok {
		idx := strings.Index(addr.String(), ":")
		return addr.String()[:idx], nil
	}
	return "-1", errors.New("not happen")
}

// Http Client

type HttpRequester struct {
	client *http.Client
}

func NewDefaultHttpRequester() HttpRequester {
	return NewHttpRequesterWithTimeout(10 * time.Second)
}

// NewHttpRequesterWithTimeout creates a http requester with specified timeout (seconds).
func NewHttpRequesterWithTimeout(timeout time.Duration) HttpRequester {
	client := &http.Client{
		Timeout: timeout,
	}
	client.Transport = &http.Transport{
		MaxIdleConns:          100,
		IdleConnTimeout:       60 * time.Second,
		ExpectContinueTimeout: 30 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}
	return HttpRequester{
		client: client,
	}
}

func NewHttpRequesterWithRoundTripper(rt http.RoundTripper) HttpRequester {
	client := &http.Client{
		Transport: rt,
	}
	return HttpRequester{
		client: client,
	}
}

func (requester HttpRequester) Get(ctx context.Context, url string, headers map[string]string) (*http.Response, []byte, error) {
	req, err := requester.createRequest(ctx, http.MethodGet, url, headers, []byte(""))
	if err != nil {
		return nil, nil, err
	}
	return requester.send(req)
}

func (requester HttpRequester) Post(ctx context.Context, url string, headers map[string]string, body []byte) (*http.Response, []byte, error) {
	req, err := requester.createRequest(ctx, http.MethodPost, url, headers, body)
	if err != nil {
		return nil, nil, err
	}
	return requester.send(req)
}

func (HttpRequester) createRequest(ctx context.Context, method, url string, headers map[string]string, body []byte) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	if len(body) > 0 {
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBuffer(body))
	} else {
		req, err = http.NewRequestWithContext(ctx, method, url, nil)
	}
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		req.Header.Add(k, v)
	}
	return req, nil
}

func (requester HttpRequester) send(req *http.Request) (*http.Response, []byte, error) {
	resp, err := requester.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, nil, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, err
	}
	return resp, body, nil
}
