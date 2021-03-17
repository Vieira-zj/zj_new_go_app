package main

import (
	"context"
	"log"
	"net"

	pb "demo.grpc/grpc.greeter/proto"
	"demo.grpc/grpc.greeter/proto/message"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	port = ":50051"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *message.HelloRequest) (*message.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &message.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("Grpc server listen at", port)

	s := grpc.NewServer()
	// for grpc web ui
	reflection.Register(s)

	pb.RegisterGreeterServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
