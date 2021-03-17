package utils

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"
)

// HTTPGet send http get request.
func HTTPGet(client http.Client, url string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("Http Req Failed " + err.Error())
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("Http Send Request Failed " + err.Error())
	}
	defer resp.Body.Close()

	content, err := ioutil.ReadAll(resp.Body) // 如果不及时从请求中获取结果，此连接会占用
	if err != nil {
		return fmt.Errorf("Read Response Body Failed " + err.Error())
	}
	fmt.Println("Response Body:", string(content))
	return nil
}

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
