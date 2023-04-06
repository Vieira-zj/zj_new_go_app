package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/html"
)

func GetHostIpAddrs() ([]string, []string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, nil, err
	}

	var (
		localIPV4    = make([]string, 0, 2)
		nonLocalIPV4 = make([]string, 0, 2)
	)
	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
			if ipNet.IP.IsLoopback() {
				localIPV4 = append(localIPV4, ipNet.IP.String())
			} else {
				nonLocalIPV4 = append(nonLocalIPV4, ipNet.IP.String())
			}
		}
	}

	return localIPV4, nonLocalIPV4, nil
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

func (requester HttpRequester) GetClient() *http.Client {
	return requester.client
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

// Parse Html Response

func GetHtmlLiValues(htmlText string) []string {
	tokens := html.NewTokenizer(strings.NewReader(htmlText))
	var (
		vals    []string
		isLiTag bool
	)
	for {
		tokenType := tokens.Next()
		switch tokenType {
		case html.ErrorToken:
			return vals
		case html.StartTagToken:
			t := tokens.Token()
			isLiTag = t.Data == "li"
		case html.TextToken:
			t := tokens.Token()
			if isLiTag {
				vals = append(vals, t.Data)
			}
			isLiTag = false
		}
	}
}
