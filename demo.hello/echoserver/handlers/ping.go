package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

var count int

// IndexHandler default index handler.
func IndexHandler(c echo.Context) error {
	return c.String(http.StatusOK, getHello())
}

// PingHandler ping handle.
func PingHandler(c echo.Context) error {
	count++
	return c.String(http.StatusOK, fmt.Sprintf("Access Count: %d", count))
}

func getHello() string {
	return "Hello, World!"
}
