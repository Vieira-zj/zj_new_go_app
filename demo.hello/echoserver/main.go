package main

import (
	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"

	"demo.hello/echoserver/handlers"
	"demo.hello/echoserver/utils"
)

func runApp() {
	// echo refer: https://echo.labstack.com/guide/request/
	deco := utils.Deco

	e := echo.New()
	e.GET("/", deco(handlers.IndexHandler))
	e.GET("/ping", utils.RateLimiterDeco(deco(handlers.PingHandler)))

	// router reg test
	e.GET("/users/", deco(handlers.Users))
	e.POST("/users/new", deco(handlers.UsersNew))
	e.GET("/users/:name", deco(handlers.UsersName))
	e.GET("/users/1/files/*", deco(handlers.UsersFiles))

	// test
	e.GET("/cover", deco(handlers.CoverHandler))
	e.GET("/sample/01", deco(handlers.SampleHandler01))

	e.Logger.SetLevel(log.INFO)
	e.Logger.Fatal(e.Start(":8081"))
}

func main() {
	runApp()
}
