package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"demo.apps/go.test/cover"
	"demo.apps/utils"
)

func main() {
	log.Println("go version:", utils.GoVersion())
	utils.SetMemoryLimit(256, utils.MB)

	httpServe(false)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	<-ctx.Done()
	cancel()

	log.Println("app exit")
}

func httpServe(isRun bool) {
	if isRun {
		router := cover.InitRouter()
		go cover.HttpServe(router)
	}
}
