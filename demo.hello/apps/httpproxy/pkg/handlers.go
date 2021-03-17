package pkg

import (
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo"
)

// PingHandler responses for ping server.
func PingHandler(c echo.Context) error {
	return c.String(http.StatusOK, "hello world")
}

// ProxyHandler proxy service.
func ProxyHandler(c echo.Context) error {
	resp, err := sendHTTPRequest(c)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}

	retBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
	}
	defer resp.Body.Close()

	headers := c.Response().Header()
	for k, v := range resp.Header {
		headers.Add(k, v[0])
	}
	headers.Add("X-Test", "TestProxy")

	c.Response().WriteHeader(http.StatusOK)
	_, err = c.Response().Writer.Write(retBody)
	return err
}
