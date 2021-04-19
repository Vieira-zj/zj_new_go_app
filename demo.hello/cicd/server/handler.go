package server

import (
	"net/http"
	"strconv"

	"demo.hello/cicd/pkg"
	"github.com/labstack/echo"
)

var (
	count   = 0
	jira    = pkg.NewJiraTool()
	treeMap = make(map[string]*pkg.JiraIssuesTree)
)

// Index .
func Index(c echo.Context) error {
	count++
	return c.String(http.StatusOK, strconv.Itoa(count))
}

// Ping .
func Ping(c echo.Context) error {
	return c.String(http.StatusOK, "ok")
}
