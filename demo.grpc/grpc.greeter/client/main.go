// Package main implements a client for Greeter service.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	pb "demo.grpc/grpc.greeter/proto"
	"demo.grpc/grpc.greeter/proto/message"
	"google.golang.org/grpc"
)

var (
	address, port, msg string
	help               bool
)

func main() {
	flag.StringVar(&address, "addr", "localhost", "Grpc client connect ip address.")
	flag.StringVar(&port, "port", "50051", "Grpc client connect port.")
	flag.StringVar(&msg, "msg", "world", "Message send to grpc server.")
	flag.BoolVar(&help, "h", false, "Help.")
	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	tmp := message.HelloRequest{}
	if b, err := json.Marshal(&tmp); err == nil {
		log.Println("default reqpuest json:", string(b))
	}

	// Set up a connection to the server.
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", address, port), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	c := pb.NewGreeterClient(conn)
	r, err := c.SayHello(ctx, &message.HelloRequest{Name: msg})
	if err != nil {
		log.Fatalf("Could not greet: %v", err)
	}
	log.Printf("Greeting: %s", r.GetMessage())
}
