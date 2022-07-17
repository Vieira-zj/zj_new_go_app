package main

import (
	"fmt"
	"time"

	"go1_1711_demo/ioc.demo/autowire_rpc/server/pkg/service/api"

	"github.com/alibaba/ioc-golang"
)

// +ioc:autowire=true
// +ioc:autowire:type=singleton

type App struct {
	ServiceStruct api.ServiceStructIOCRPCClient `rpc-client:",address=127.0.0.1:2022"`
}

func (a *App) Run() {
	for {
		time.Sleep(time.Second * 3)
		user, err := a.ServiceStruct.GetUser("laurence", 23)
		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("get user = %+v\n", user)

	}
}

func main() {
	if err := ioc.Load(); err != nil {
		panic(err)
	}

	app, err := GetAppSingleton()
	if err != nil {
		panic(err)
	}
	app.Run()
}
