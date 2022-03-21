package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/labstack/echo"
)

var count int32

// IndexHandler default index handler.
func IndexHandler(c echo.Context) error {
	return c.String(http.StatusOK, getHello())
}

// PingHandler ping handle.
func PingHandler(c echo.Context) error {
	ret := atomic.AddInt32(&count, 1)
	return c.String(http.StatusOK, fmt.Sprintf("Access Count: %d", ret))
}

type requestMeta struct {
	Host   string `json:"host"`
	URI    string `json:"uri"`
	Method string `json:"method"`
}

type requestData struct {
	Meta    requestMeta       `json:"meta"`
	Headers map[string]string `json:"headers"`
	Body    interface{}       `json:"body"`
}

// MirrorHandler returns copied request data for test.
func MirrorHandler(c echo.Context) error {
	req := c.Request()
	body := req.Body
	defer body.Close()

	respBody, err := buildRequestBody(body)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

	reqData := requestData{
		Meta: requestMeta{
			Host:   req.Host,
			URI:    req.URL.RequestURI(),
			Method: req.Method,
		},
		Headers: formatRequestHeaders(req.Header),
		Body:    respBody,
	}
	return c.JSON(http.StatusOK, reqData)
}

func getHello() string {
	return "Hello, World!"
}

func formatRequestHeaders(headers map[string][]string) map[string]string {
	newHeaders := make(map[string]string, len(headers))
	for k, v := range headers {
		newHeaders[k] = strings.Join(v, ",")
	}
	return newHeaders
}

func buildRequestBody(body io.ReadCloser) (interface{}, error) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return nil, err
	}

	if len(b) == 0 {
		return "", nil
	}
	if b[0] != '{' {
		return string(b), nil
	}

	var respBody interface{}
	if err := json.Unmarshal(b, &respBody); err != nil {
		return nil, err
	}
	return respBody, nil
}
