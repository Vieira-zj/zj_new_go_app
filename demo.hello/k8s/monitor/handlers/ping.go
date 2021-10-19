package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

var count int

// Home .
func Home(c echo.Context) error {
	count++
	return c.String(http.StatusOK, fmt.Sprintf("Access Count: %d", count))
}

// Ping .
func Ping(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
