package handlers

import (
	"fmt"
	"net/http"
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

func getHello() string {
	return "Hello, World!"
}
