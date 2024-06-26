package application

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"demo.grpc/grpc.impl/pkg/interceptor"
	"demo.grpc/grpc.impl/pkg/protoc"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

var mockData = map[string]string{
	"/account.DepositService/Deposit":       `{"ok": true}`,
	"/account.DepositService/CreateAccount": `{"return_code": 900100}`,
	"/greeter.Greeter/SayHello":             `{"content": "from grpc.app"}`,
}

func RunGrpcServer(port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return err
	}

	grpcServer := newGrpcServer()
	go func() {
		log.Println("grpc server listen at:", port)
		if err := grpcServer.Serve(lis); err != nil {
			log.Println("grpc server error:", err)
		}
		grpcServer.GracefulStop()
	}()
	return nil
}

func newGrpcServer() *grpc.Server {
	interceptors := []grpc.StreamServerInterceptor{
		interceptor.RecoverStreamServerInterceptor(),
		interceptor.LoggingStreamServerInterceptor(),
	}

	serverOpts := []grpc.ServerOption{
		// use encoding.RegisterCodec instead
		// grpc.CustomCodec(&codec.ProtoJson{}),
		grpc.MaxRecvMsgSize(8 * 1024 * 1024),
		grpc.UnknownServiceHandler(unKnownSsHandler),
		grpc_middleware.WithStreamServerChain(interceptors...),
	}
	return grpc.NewServer(serverOpts...)
}

func unKnownSsHandler(srv interface{}, serverStream grpc.ServerStream) error {
	method, ok := grpc.MethodFromServerStream(serverStream)
	if !ok {
		return fmt.Errorf("get method from server stream failed")
	}
	log.Println(strings.Repeat("*", 10), "process:", method)

	// handle metadata
	md, ok := metadata.FromIncomingContext(serverStream.Context())
	if !ok {
		return fmt.Errorf("get metadata from incoming context failed")
	}
	for k, v := range md {
		log.Printf("metadata: key=%v,value=%v", k, v)
	}

	// handle request
	coder := GetProtoCoder()
	req, err := coder.NewReqProtoMessage(method)
	if err != nil {
		return fmt.Errorf("new req proto msg error: %v", err)
	}
	if err := serverStream.RecvMsg(req); err != nil {
		return err
	}
	log.Println("receive msg:", req.String())

	// handle response
	respBody, ok := mockData[method] // mock 逻辑
	if !ok {
		return fmt.Errorf("no matched resp found")
	}
	resp, err := coder.BuildRespProtoMessage(method, respBody)
	if err != nil {
		return fmt.Errorf("build resp proto msg error: %v", err)
	}
	log.Println("send msg:", resp.String())
	return serverStream.SendMsg(resp)
}

// ProtoCoder

var (
	protoCoder     protoc.Coder
	protoCoderOnce sync.Once
)

func GetProtoCoder() protoc.Coder {
	protoCoderOnce.Do(func() {
		mds, err := loadProto()
		if err != nil {
			log.Fatal(err)
		}
		for k, md := range mds {
			log.Println("load:", k, md.GetName())
		}
		protoCoder = protoc.NewCoder(mds)
	})
	return protoCoder
}

func loadProto() (map[string]*desc.MethodDescriptor, error) {
	path := filepath.Join(os.Getenv("PROJECT_ROOT"), "grpc.app/proto")
	dirPaths, err := protoc.GetAllProtoDirs(path)
	if err != nil {
		return nil, err
	}
	return protoc.LoadProtoFiles(dirPaths...)
}
