package server

import (
	"net/http"
	"strconv"
	"sync/atomic"

	"github.com/labstack/echo"
)

var count int32

// Index .
func Index(c echo.Context) error {
	ret := atomic.AddInt32(&count, 1)
	return c.String(http.StatusOK, strconv.Itoa(int(ret)))
}

// Ping .
func Ping(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
