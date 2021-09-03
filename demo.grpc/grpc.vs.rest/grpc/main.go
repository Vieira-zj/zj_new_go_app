package main

import (
	"context"
	"log"
	"net"

	pb "demo.grpc/grpc.vs.rest/proto"
	"google.golang.org/grpc"
)

type server struct {
}

var count int

func (s *server) DoSomething(_ context.Context, random *pb.Random) (*pb.Random, error) {
	count++
	log.Printf("grpc access count: %d", count)
	random.RandomString = "[Updated] " + random.RandomString
	return random, nil
}

func main() {
	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterRandomServiceServer(s, &server{})
	log.Println("grpc server start listen at: 9090")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
