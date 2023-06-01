package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"time"
)

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

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return resp, nil, err
	}
	return resp, body, nil
}
