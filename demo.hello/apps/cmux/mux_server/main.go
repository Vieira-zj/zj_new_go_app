package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

//
// Desc: 使用 cmux 实现服务端连接多路复用
//
// Test http:
// curl -i http://localhost:50051
//
// Test grpc:
// ./greeter_client -name=foo
//

func main() {
	lis, err := net.Listen("tcp", "localhost:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	mux := cmux.New(lis)

	// gRPC match rule
	grpcLis := mux.MatchWithWriters(
		cmux.HTTP2MatchHeaderFieldSendSettings("content-type", "application/grpc"))
	// otherwise serve http
	httpLis := mux.Match(cmux.Any())

	// gRPC server
	grpcSvr := grpc.NewServer()
	pb.RegisterGreeterServer(grpcSvr, &server{})

	// HTTP server
	httpSvr := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("greet from HTTP\n"))
		}),
	}

	go grpcSvr.Serve(grpcLis)
	go httpSvr.Serve(httpLis)

	log.Printf("serve both grpc and http at [%v]", lis.Addr())
	if err := mux.Serve(); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

type server struct {
	pb.UnimplementedGreeterServer
}

func (*server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	return &pb.HelloReply{
		Message: fmt.Sprintf("Hello %s, greet from gRPC", in.GetName()),
	}, nil
}
