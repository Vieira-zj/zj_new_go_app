package utils

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"
)

/*
HTTP
*/

// HTTPUtils a http client utils.
type HTTPUtils struct {
	client http.Client
}

// NewHTTPUtils creates a http utils instance.
func NewHTTPUtils(isKeepAlive bool) *HTTPUtils {
	if !isKeepAlive {
		return &HTTPUtils{
			client: http.Client{},
		}
	}

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
	return &HTTPUtils{
		client: http.Client{Transport: httpTransport},
	}
}

// Get sends http get request.
func (utils *HTTPUtils) Get(ctx context.Context, url string, headers map[string]string) ([]byte, error) {
	req, err := utils.createRequest(ctx, http.MethodGet, url, headers, "")
	if err != nil {
		return nil, err
	}
	return utils.send(req)
}

// GetWithAuth sends http get request with auth enabled.
func (utils *HTTPUtils) GetWithAuth(ctx context.Context, url string, headers map[string]string, name, password string) ([]byte, error) {
	req, err := utils.createRequest(ctx, http.MethodGet, url, headers, "")
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(name, password)
	return utils.send(req)
}

// Post sends http post request.
func (utils *HTTPUtils) Post(ctx context.Context, url string, headers map[string]string, body string) ([]byte, error) {
	req, err := utils.createRequest(ctx, http.MethodPost, url, headers, body)
	if err != nil {
		return nil, err
	}
	return utils.send(req)
}

// PostWithAuth sends http post request with auth enabled.
func (utils *HTTPUtils) PostWithAuth(ctx context.Context, url string, headers map[string]string, body, name, password string) ([]byte, error) {
	req, err := utils.createRequest(ctx, http.MethodPost, url, headers, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth(name, password)
	return utils.send(req)
}

func (utils *HTTPUtils) createRequest(ctx context.Context, method string, url string, headers map[string]string, body string) (*http.Request, error) {
	var (
		req *http.Request
		err error
	)
	if len(body) > 0 {
		req, err = http.NewRequestWithContext(ctx, method, url, strings.NewReader(body))
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

func (utils *HTTPUtils) send(req *http.Request) ([]byte, error) {
	resp, err := utils.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// 如果不及时从请求中获取结果，此连接会占用
	return ioutil.ReadAll(resp.Body)
}

/*
TCP
*/

// PingTCP pings a tcp connection.
func PingTCP(host string, port string) (bool, error) {
	addr := net.JoinHostPort(host, port)
	timeout := time.Duration(3) * time.Second
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false, err
	}

	if conn != nil {
		defer conn.Close()
		fmt.Println("tcp opened:", addr)
		return true, nil
	}
	return false, errors.New("tcp connection is nil")
}
