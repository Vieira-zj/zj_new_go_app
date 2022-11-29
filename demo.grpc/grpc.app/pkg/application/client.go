package application

import (
	"context"
	"fmt"
	"strconv"
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
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	md := &metadata.MD{}
	opts = append(opts, grpc.Trailer(md))

	var key GrpcRetCodeMsg = "code_msg"
	ctx = context.WithValue(ctx, key, md)

	conn, err := createGrpcClientConn(ctx, target)
	if err != nil {
		return err
	}
	if err = conn.Invoke(ctx, fullMethod, req, resp, opts...); err != nil {
		return err
	}
	code, msg := extractErrFromMetadata(md)
	if code != 0 {
		return fmt.Errorf("ret_code=%d, msg=%s", code, msg)
	}
	return nil
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

func extractErrFromMetadata(md *metadata.MD) (code int32, msg string) {
	if c := md.Get("code"); len(c) > 0 {
		codeStr, _ := strconv.Atoi(c[0])
		code = int32(codeStr)
	}
	if m := md.Get("msg"); len(m) > 0 {
		msg = m[0]
	}
	return
}
