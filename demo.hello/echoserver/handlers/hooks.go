package handlers

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/labstack/echo"
)

// Deco decorate echo function.
func Deco(fn func(echo.Context) error) func(echo.Context) error {
	return func(c echo.Context) error {
		preHook(c)
		err := fn(c)
		afterHook(c)
		return err
	}
}

func preHook(c echo.Context) {
	printRequestInfo(c)
}

func afterHook(c echo.Context) {
	// TODO:
}

func printRequestInfo(c echo.Context) {
	printDivLine(c)
	request := c.Request()
	c.Logger().Info("| Host: ", request.Host)
	c.Logger().Info("| Url: ", request.URL)
	c.Logger().Info("| Method: ", request.Method)
	printHeaders(c, request.Header)

	content, err := ioutil.ReadAll(request.Body)
	if err != nil {
		c.Logger().Error(err)
	}
	if len(content) > 0 {
		c.Logger().Info("| Body: ", string(content))
	}
	printDivLine(c)
}

func printDivLine(c echo.Context) {
	c.Logger().Info("| " + strings.Repeat("*", 60))
}

func printHeaders(c echo.Context, headers map[string][]string) {
	c.Logger().Info("| Headers:")
	for k, v := range headers {
		c.Logger().Info(fmt.Sprintf("|   %s: %s\n", k, v[0]))
	}
}
