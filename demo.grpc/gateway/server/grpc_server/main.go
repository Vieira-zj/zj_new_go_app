package main

import (
	"context"
	"log"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"time"

	pb "demo.grpc/gateway/proto/demo/hello"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/reflection"
)

const (
	port        = ":9090"
	profilePort = ":8024"
	gocPort     = ":8026"
)

var (
	isPProf = false
)

type server1 struct {
	pb.UnimplementedService1Server
}

type server2 struct {
	pb.UnimplementedService1Server
}

func (s *server1) Echo(ctx context.Context, in *pb.StringMessage) (*pb.StringMessage, error) {
	log.Printf("Received: %v", in.GetValue())
	time.Sleep(time.Duration(rand.Int31n(300)) * time.Millisecond) // mock profile
	return &pb.StringMessage{Value: "Echo " + in.GetValue()}, nil
}

func (s *server2) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		log.Println("get metadata error")
	}
	log.Println("income metadata:")
	log.Println(md)

	// by custom mather
	if uid, ok := md["x-user-id"]; ok {
		log.Println("user-id from metadata: " + uid[0])
	}
	// by default matcher "Grpc-Metadata-Uid"
	if uid, ok := md["uid"]; ok {
		log.Println("uid from metadata: " + uid[0])
	}

	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	runtime.SetBlockProfileRate(1)

	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	} else {
		log.Println("grpc listen at: " + port)
	}
	defer lis.Close()

	if isPProf {
		go func() {
			log.Println("start pprof server at:", profilePort)
			if err := http.ListenAndServe(profilePort, nil); err != nil {
				log.Fatalf("fail to start pprof server: %v\n", err)
			}
		}()
	}

	s := grpc.NewServer()
	defer s.GracefulStop()

	pb.RegisterService1Server(s, &server1{})
	pb.RegisterService2Server(s, &server2{})

	// for grpc web ui
	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
