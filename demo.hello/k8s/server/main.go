package main

import (
	"demo.hello/k8s/server/handlers"
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"
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
	// TODO:
}

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	e.GET("/", deco(handlers.Home))
	e.GET("/ping", deco(handlers.Ping))

	e.GET("/monitor/pods", deco(handlers.GetPodsStatus))

	e.Logger.Info("Start cluster monitor at 8081.")
	e.Logger.Fatal(e.Start(":8081"))
}
