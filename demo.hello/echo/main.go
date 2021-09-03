package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo"
	"github.com/labstack/gommon/log"

	"demo.hello/echo/handlers"
	"demo.hello/echo/utils"
)

func main() {
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

	// data
	e.GET("/data/rowspan", deco(handlers.GetTableRowSpanData))

	// test
	e.GET("/cover", deco(handlers.CoverHandler))
	e.GET("/sample/01", deco(handlers.SampleHandler01))
	e.GET("/sample/02", deco(handlers.SampleHandler02))

	go func() {
		e.Logger.SetLevel(log.INFO)
		e.Logger.Fatal(e.Start(":8081"))
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)

	<-quit
	e.Logger.Info("Stopping server.")
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		panic(err)
	}
}
