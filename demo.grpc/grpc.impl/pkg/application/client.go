package application

import (
	"context"
	"log"
	"time"

	"demo.grpc/grpc.app/pkg/codec"
	"demo.grpc/grpc.app/pkg/interceptor"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type GrpcRetCodeMsg string

// GrpcCallWithJson calls rpc with json codec, the request and reply are both json string.
func GrpcCallWithJson(ctx context.Context, target, fullMethod string, req []byte, opts ...grpc.CallOption) ([]byte, error) {
	jsonOpts := []grpc.CallOption{
		grpc.ForceCodec(&codec.JsonFrame{}),
		codec.WithJsonCodec(),
	}
	opts = append(opts, jsonOpts...)

	if len(req) == 0 {
		req = []byte("{}")
	}
	request := &codec.JsonFrame{RawData: req}
	resp := &codec.JsonFrame{}

	err := GrpcCall(ctx, target, fullMethod, request, resp, opts...)
	return resp.RawData, err
}

func GrpcCall(ctx context.Context, target, fullMethod string, req interface{}, resp interface{}, opts ...grpc.CallOption) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	conn, err := getGrpcClientConn(ctx, target)
	if err != nil {
		return err
	}

	md := metadata.MD{}
	md.Set("msg", "grpc.app client")
	newCtx := metadata.NewOutgoingContext(ctx, md)
	return conn.Invoke(newCtx, fullMethod, req, resp, opts...)
}

// grpc conn pool

var connPool = make(map[string]*grpc.ClientConn, 4)

func getGrpcClientConn(ctx context.Context, target string) (*grpc.ClientConn, error) {
	if conn, ok := connPool[target]; ok {
		log.Println("reuse conn for:", target)
		return conn, nil
	}

	conn, err := createGrpcClientConn(ctx, target)
	if err != nil {
		return nil, err
	}
	connPool[target] = conn
	return conn, nil
}

func createGrpcClientConn(ctx context.Context, target string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
		grpc.WithUnaryInterceptor(grpc_middleware.ChainUnaryClient(
			interceptor.LoggingClientInterceptor(),
			interceptor.RetryClientInterceptor(),
		)),
	}
	return grpc.DialContext(ctx, target, opts...)
}
