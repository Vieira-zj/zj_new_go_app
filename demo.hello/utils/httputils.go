package utils

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	neturl "net/url"
	"sync"
	"time"
)

/*
Common
*/

// BuildURL .
func BuildURL(host string, query map[string]string) (string, error) {
	u, err := neturl.Parse(host)
	if err != nil {
		return "", nil
	}

	q := u.Query()
	for k, v := range query {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String(), nil
}

// IsNetworkError if network error, we should retry http reqeust.
func IsNetworkError(err error) bool {
	if err == io.EOF {
		return true
	}
	_, ok := err.(net.Error)
	return ok
}

// IsTemporaryNetworkError .
func IsTemporaryNetworkError(err error) bool {
	if nerr, ok := err.(net.Error); ok && nerr.Temporary() {
		return true
	}
	if operr, ok := err.(*net.OpError); ok && operr.Temporary() {
		return true
	}
	return false
}

// GetLocalHostIPs gets local host ip address.
func GetLocalHostIPs() (hosts []string, err error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok && ipNet.IP.To4() != nil {
			hosts = append(hosts, ipNet.IP.String())
		}
	}
	return
}

// GetRemoteHostIPs .
func GetRemoteHostIPs(url string) ([]string, error) {
	parsed, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}

	ips, err := net.LookupIP(parsed.Host)
	if err != nil {
		return nil, err
	}

	ret := make([]string, 0, len(ips))
	for _, ip := range ips {
		ret = append(ret, ip.To4().String())
	}
	return ret, nil
}

/*
HTTP
*/

var (
	client   *http.Client
	httpOnce sync.Once
)

// HTTPUtils a http client utils.
type HTTPUtils struct {
	client *http.Client
}

// NewDefaultHTTPUtils creates a http util with default client.
func NewDefaultHTTPUtils() *HTTPUtils {
	httpOnce.Do(func() {
		client = &http.Client{
			Timeout: 10 * time.Second,
		}
		client.Transport = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
	})

	return &HTTPUtils{
		client: client,
	}
}

// NewHTTPUtils creates a http util instance.
func NewHTTPUtils(isKeepAlives bool) *HTTPUtils {
	httpOnce.Do(func() {
		// http client 默认是长链接
		var httpTransport *http.Transport
		if isKeepAlives {
			httpTransport = &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 60 * time.Second,
				}).DialContext,
				MaxIdleConns:          500,              // 最大空闲连接
				IdleConnTimeout:       60 * time.Second, // 空闲连接的超时时间
				ExpectContinueTimeout: 30 * time.Second, // 等待服务第一个响应的超时时间
				MaxIdleConnsPerHost:   100,              // 每个host保持的空闲连接数
			}
		} else {
			httpTransport = &http.Transport{
				DisableKeepAlives: true,
			}
		}
		httpTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
		client = &http.Client{Transport: httpTransport}
	})

	return &HTTPUtils{
		client: client,
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
		req, err = http.NewRequestWithContext(ctx, method, url, bytes.NewBufferString(body))
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
	if resp != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, err
	}

	// 如果不及时从请求中获取结果，此连接会占用
	return ioutil.ReadAll(resp.Body)
}

// GetV2 sends http get request.
func (utils *HTTPUtils) GetV2(ctx context.Context, url string, headers map[string]string) (*http.Response, []byte, error) {
	req, err := utils.createRequest(ctx, http.MethodGet, url, headers, "")
	if err != nil {
		return nil, nil, err
	}
	return utils.sendV2(req)
}

// PostV2 sends http post request.
func (utils *HTTPUtils) PostV2(ctx context.Context, url string, headers map[string]string, body string) (*http.Response, []byte, error) {
	req, err := utils.createRequest(ctx, http.MethodPost, url, headers, body)
	if err != nil {
		return nil, nil, err
	}
	return utils.sendV2(req)
}

// sendV2 returns both http response and resp body.
func (utils *HTTPUtils) sendV2(req *http.Request) (*http.Response, []byte, error) {
	resp, err := utils.client.Do(req)
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

// HijackHTTPRequest returns http conntion by hijack.
func HijackHTTPRequest(w http.ResponseWriter) (net.Conn, *bufio.ReadWriter, error) {
	hj, ok := w.(http.Hijacker)
	if !ok {
		err := errors.New("http server don't support hijacking")
		return nil, nil, err
	}
	return hj.Hijack()
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

/*
CORS
*/

// AddCorsHeadersForOptions adds cors access control allow headers for options request.
func AddCorsHeadersForOptions(w http.ResponseWriter) {
	AddCorsHeadersForSimple(w)
	w.Header().Add("Access-Control-Allow-Headers", "Accept,Origin,Content-Type,Authorization")
	w.Header().Add("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
}

// AddCorsHeadersForSimple adds cors access control allow headers for simple request.
func AddCorsHeadersForSimple(w http.ResponseWriter) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Credentials", "true")
}
