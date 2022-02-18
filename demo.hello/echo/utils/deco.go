package utils

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/labstack/echo"
)

var ipRateLimiter = NewIPRateLimiter(1, 5)

// Deco log decorate function.
func Deco(fn func(echo.Context) error) func(echo.Context) error {
	return func(c echo.Context) error {
		preHook(c)
		err := fn(c)
		afterHook(c)
		return err
	}
}

// RateLimiterDeco rate limiter decorate functions.
func RateLimiterDeco(fn func(echo.Context) error) func(echo.Context) error {
	return func(c echo.Context) error {
		l := ipRateLimiter.GetLimiter(c.Request().Host)
		if !l.Allow() {
			return c.String(http.StatusTooManyRequests, http.StatusText(http.StatusTooManyRequests))
		}
		return fn(c)
	}
}

/*
Logs
*/

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

	body := request.Body
	defer body.Close()
	content, err := ioutil.ReadAll(body)
	if err != nil {
		c.Logger().Error(err)
	}
	if len(content) > 0 {
		c.Logger().Info("| Body: ", string(content))
	}
	c.Set("req_body", content)
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
