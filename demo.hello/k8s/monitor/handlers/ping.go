package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo"
)

var count int

// Home / handler.
func Home(c echo.Context) error {
	return c.String(http.StatusOK, "k8s monitor tool.")
}

// Ping /ping handler.
func Ping(c echo.Context) error {
	count++
	return c.String(http.StatusOK, fmt.Sprintf("Access Count: %d", count))
}
