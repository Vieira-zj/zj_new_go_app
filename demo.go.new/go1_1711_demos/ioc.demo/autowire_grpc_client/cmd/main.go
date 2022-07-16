package main

import (
	"context"
	"fmt"
	"time"

	"github.com/alibaba/ioc-golang/config"

	"go1_1711_demo/ioc.demo/autowire_grpc_client/api"
	"go1_1711_demo/ioc.demo/autowire_grpc_client/cmd/service1"
	"go1_1711_demo/ioc.demo/autowire_grpc_client/cmd/service2"
	"go1_1711_demo/ioc.demo/autowire_grpc_client/cmd/struct1"

	googleGRPC "google.golang.org/grpc"

	"github.com/alibaba/ioc-golang"
	"github.com/alibaba/ioc-golang/autowire"
	"github.com/alibaba/ioc-golang/extension/autowire/grpc"
)

func init() {
	// register grpc client
	grpc.RegisterStructDescriptor(&autowire.StructDescriptor{
		Factory: func() interface{} {
			return new(api.HelloServiceClient)
		},
		ParamFactory: func() interface{} {
			return &googleGRPC.ClientConn{}
		},
		ConstructFunc: func(impl interface{}, param interface{}) (interface{}, error) {
			conn := param.(*googleGRPC.ClientConn)
			fmt.Println("create conn target ", conn.Target())
			return api.NewHelloServiceClient(conn), nil
		},
	})
}

// +ioc:autowire=true
// +ioc:autowire:type=singleton
// +ioc:autowire:alias=App

type App struct {
	HelloServiceClient api.HelloServiceClient `grpc:"hello-service"`

	ExampleService1Impl1 service1.Service1 `singleton:"go1_1711_demo/ioc.demo/autowire_grpc_client/cmd/service1.Impl1"`
	ExampleService2Impl1 service2.Service2 `singleton:"go1_1711_demo/ioc.demo/autowire_grpc_client/cmd/service2.Impl1"`
	ExampleService2Impl2 service2.Service2 `singleton:"go1_1711_demo/ioc.demo/autowire_grpc_client/cmd/service2.Impl2"`

	ExampleStruct1 *struct1.Struct1 `singleton:""`
}

func (a *App) Run() {
	for {
		time.Sleep(3 * time.Second)
		name := "laurence"
		rsp, err := a.HelloServiceClient.SayHello(context.Background(), &api.HelloRequest{
			Name: name,
		})
		if err != nil {
			panic(err)
		}

		fmt.Println("App call grpc get: " + rsp.Reply)

		fmt.Println("ExampleService1Impl1 call grpc get:" + a.ExampleService1Impl1.Hello(name+"_service1_impl1"))
		fmt.Println("ExampleService2Impl1 call grpc get:" + a.ExampleService2Impl1.Hello(name+"_service2_impl1"))
		fmt.Println("ExampleService2Impl2 call grpc get:" + a.ExampleService2Impl2.Hello(name+"_service2_impl2"))
		fmt.Println("ExampleStruct1 call grpc get:" + a.ExampleStruct1.Hello(name+"_struct"))
	}
}

func main() {
	if err := ioc.Load(
		config.WithSearchPath("../conf"),
		config.WithConfigName("ioc_golang")); err != nil {
		panic(err)
	}

	app, err := GetAppSingleton()
	if err != nil {
		panic(err)
	}
	app.Run()
}
