package main

import (
	"os"
	"os/signal"
	"syscall"

	"demo.hello/apps/tcpserver/pkg"
)

func main() {
	closeCh := make(chan struct{})
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGKILL, syscall.SIGINT)

	go func() {
		<-sigCh
		closeCh <- struct{}{}
	}()

	pkg.ListenAndServe(":8080", pkg.NewEchoHandler(), closeCh)
}
