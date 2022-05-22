package main

import (
	"context"
	"log"
	"net"
	"time"

	pb "demo.grpc/grpc.mock/demo_protobuf"
	"google.golang.org/grpc"
)

func main() {
	conn, err := grpc.Dial(net.JoinHostPort("127.0.0.1", "50051"), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	c := pb.NewExampleServiceClient(conn)
	r, err := c.ExampleMethod(ctx, &pb.ExampleRequest{
		// Two: "grpcmock",
		Two: "zzzz",
	})
	if err != nil {
		log.Fatalf("Call method error: %v", err)
	}
	log.Printf("Response: code=%d, msg=%s", r.GetCode(), r.GetMessage())
}
