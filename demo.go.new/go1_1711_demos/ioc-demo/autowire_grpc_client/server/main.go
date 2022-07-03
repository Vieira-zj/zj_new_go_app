package main

import (
	"context"
	"net"

	"go1_1711_demo/ioc-demo/autowire_grpc_client/api"

	"google.golang.org/grpc"
)

type HelloServiceImpl struct {
	api.UnimplementedHelloServiceServer
}

func (h *HelloServiceImpl) SayHello(_ context.Context, req *api.HelloRequest) (*api.HelloResponse, error) {
	return &api.HelloResponse{
		Reply: "Hello " + req.Name,
	}, nil
}

func main() {
	lst, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}

	grpcServer := grpc.NewServer()
	grpcServer.RegisterService(&api.HelloService_ServiceDesc, &HelloServiceImpl{})
	if err := grpcServer.Serve(lst); err != nil {
		panic(err)
	}
}
