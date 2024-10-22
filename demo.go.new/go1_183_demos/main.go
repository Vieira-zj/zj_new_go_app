package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"demo.apps/go.test/cover"
	"demo.apps/utils"
)

func main() {
	log.Println("go version:", utils.GoVersion())
	utils.SetMemoryLimit(256, utils.MB)

	httpServe(false)

	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall. SIGKILL but can"t be catch, so don't need add it
	ctx, cancel := signal.NotifyContext(context.TODO(), syscall.SIGTERM, syscall.SIGINT)
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
