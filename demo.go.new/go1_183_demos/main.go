package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"

	"demo.apps/go.test/cover"
)

func main() {
	ver := runtime.Version()
	fmt.Println("go version:", ver)

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
