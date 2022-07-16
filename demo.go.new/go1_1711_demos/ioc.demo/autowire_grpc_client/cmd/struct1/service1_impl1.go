package struct1

import (
	"context"

	"go1_1711_demo/ioc.demo/autowire_grpc_client/api"
)

// +ioc:autowire=true
// +ioc:autowire:type=singleton

type Struct1 struct {
	HelloServiceClient api.HelloServiceClient `grpc:"hello-service"`
}

func (i *Struct1) Hello(name string) string {
	rsp, err := i.HelloServiceClient.SayHello(context.Background(), &api.HelloRequest{
		Name: name,
	})
	if err != nil {
		panic(err)
	}
	return rsp.Reply
}
