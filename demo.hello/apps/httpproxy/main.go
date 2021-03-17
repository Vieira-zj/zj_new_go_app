package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"

	"demo.hello/apps/httpproxy/pkg"
)

var targetHost, port string

func startServer() {
	e := echo.New()
	e.GET("/ping", deco(pkg.PingHandler))
	e.Any("/*", deco(pkg.ProxyHandler))

	e.Logger.SetLevel(log.DEBUG)
	e.Logger.Fatal(e.Start(":" + port))
}

func main() {
	flag.StringVar(&targetHost, "t", "127.0.0.1:8081", "target service host ip:port.")
	flag.StringVar(&port, "p", "8088", "proxy server listen port.")
	help := flag.Bool("h", false, "help.")
	flag.Parse()

	if *help {
		flag.Usage()
		return
	}
	startServer()
}

func deco(fn func(echo.Context) error) func(echo.Context) error {
	return func(c echo.Context) error {
		preHook(c)
		err := fn(c)
		afterHook(c)
		return err
	}
}

func preHook(c echo.Context) {
	header := c.Request().Header
	header.Add("X-Test", "TestProxy")
	header.Add("X-Target", targetHost)

	logRequestInfo(c)
}

func afterHook(c echo.Context) {
}

func logRequestInfo(c echo.Context) {
	logDivLine(c)
	request := c.Request()
	c.Logger().Info("| Host: ", request.Host)
	c.Logger().Info("| Path: ", request.RequestURI)
	c.Logger().Info("| Method: ", request.Method)
	logHeaders(c, request.Header)
	logDivLine(c)
}

func logDivLine(c echo.Context) {
	c.Logger().Info("| " + strings.Repeat("*", 60))
}

func logHeaders(c echo.Context, headers http.Header) {
	c.Logger().Info("| Headers:")
	for k, v := range headers {
		c.Logger().Info(fmt.Sprintf("|   %s: %s\n", k, v[0]))
	}
}
