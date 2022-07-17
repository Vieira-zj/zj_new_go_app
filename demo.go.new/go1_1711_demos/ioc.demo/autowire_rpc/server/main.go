package main

import (
	_ "go1_1711_demo/ioc.demo/autowire_rpc/server/pkg/service"

	"github.com/alibaba/ioc-golang"
)

func main() {
	if err := ioc.Load(); err != nil {
		panic(err)
	}
	select {}
}
