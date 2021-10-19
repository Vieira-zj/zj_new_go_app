package main

import (
	"flag"
	"fmt"

	"demo.hello/k8s/monitor/handlers"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
)

var (
	addr string
	help bool
)

func deco(fn func(echo.Context) error) func(echo.Context) error {
	return func(c echo.Context) error {
		preHook(c)
		err := fn(c)
		afterHook(c)
		return err
	}
}

func preHook(c echo.Context) {
	request := c.Request()
	c.Logger().Info("| Host: ", request.Host)
	c.Logger().Info("| Url: ", request.URL)
	c.Logger().Info("| Method: ", request.Method)
}

func afterHook(c echo.Context) {
}

func main() {
	flag.StringVar(&addr, "addr", "8081", "server listen port.")
	flag.BoolVar(&help, "h", false, "help.")
	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	e.GET("/", deco(handlers.Home))
	e.GET("/ping", deco(handlers.Ping))

	e.GET("/monitor/pods", deco(handlers.GetPodsStatus))

	addr = fmt.Sprintf(":%s" + addr)
	e.Logger.Info("Start cluster monitor at " + addr)
	e.Logger.Fatal(e.Start(addr))
}
