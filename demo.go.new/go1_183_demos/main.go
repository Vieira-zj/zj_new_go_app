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

	httpServe()

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	<-ctx.Done()
	cancel()

	log.Println("app exit")
}

func httpServe() {
	router := cover.InitRouter()
	go cover.HttpServe(router)
}
