package handlers

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

var (
	count int
	// IsNotify .
	IsNotify = true
)

// Index .
func Index(c echo.Context) error {
	count++
	return c.String(http.StatusOK, fmt.Sprintf("Access Count: %d", count))
}

// Ping .
func Ping(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}

// SetNotify .
func SetNotify(c echo.Context) error {
	isNotify := c.QueryParam("open")
	if strings.ToLower(isNotify) == "true" {
		IsNotify = true
	} else {
		IsNotify = false
	}
	return c.String(http.StatusOK, "set success")
}
