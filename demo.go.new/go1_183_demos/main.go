package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"runtime"
	"runtime/debug"
	"time"

	"demo.apps/go.test/cover"
)

var (
	KB = int64(1 << 10)
	MB = int64(1 << 20)
	GB = int64(1 << 30)
)

func main() {
	showGoVersion()
	setMemoryLimit(256, MB)

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

// Sys Utils

func showGoVersion() {
	ver := runtime.Version()
	log.Println("go version:", ver)
}

// refer: https://pkg.go.dev/runtime/debug#SetMemoryLimit
func setMemoryLimit(size, unit int64) {
	pre := debug.SetMemoryLimit(size * unit)
	log.Printf("mem limit: pre=%dMB, cur=%dMB", pre/MB, size)
}

//nolint:unused
func forceFreeMemory() {
	debug.FreeOSMemory()
}

//nolint:unused
func collectGCState(ctx context.Context, interval time.Duration) {
	t := time.NewTicker(interval)
	defer t.Stop()

	state := debug.GCStats{}
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				debug.ReadGCStats(&state)
				log.Printf("%+v", state)
			}
		}
	}()
}
