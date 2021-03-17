package main

import (
	"log"
	"net"

	"demo.grpc/grpc.unittest/internal/account"
	pb "demo.grpc/grpc.unittest/proto/account"
	"google.golang.org/grpc"
)

// Refer: http://www.inanzzz.com/index.php/post/w9qr/unit-testing-golang-grpc-client-and-server-application-with-bufconn-package

func main() {
	log.Println("Grpc Server running ...")

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer()
	pb.RegisterDepositServiceServer(server, &account.DepositServer{})
	log.Fatal(server.Serve(listener))
}
