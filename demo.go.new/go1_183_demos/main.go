package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"

	"demo.apps/go.test/cover"
)

var (
	KB = int64(1 << 10)
	MB = int64(1 << 20)
	GB = int64(1 << 30)
)

func main() {
	ver := runtime.Version()
	fmt.Println("go version:", ver)

	setMemoryLimit(256)

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

// refer: https://pkg.go.dev/runtime/debug#SetMemoryLimit
func setMemoryLimit(size int64) {
	pre := debug.SetMemoryLimit(size * MB)
	fmt.Printf("mem limit: pre=%dMB, cur=%dMB\n", pre/MB, size)
}
