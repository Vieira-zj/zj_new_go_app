package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/grpc"
	md "google.golang.org/grpc/metadata"
)

type GrpcRetCodeMsg string

func GrpcCall(ctx context.Context, target, fullMethod string, req interface{}, resp interface{}) error {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	m := &md.MD{}
	opts := []grpc.CallOption{
		grpc.Trailer(m),
	}

	var key GrpcRetCodeMsg = "code_msg"
	ctx = context.WithValue(ctx, key, m)

	conn, err := CreateGrpcConnection(ctx, target)
	if err != nil {
		return err
	}
	if err = conn.Invoke(ctx, fullMethod, req, resp, opts...); err != nil {
		return err
	}
	code, msg := extractError(m)
	if code != 0 {
		return fmt.Errorf("ret_code=%d, msg=%s", code, msg)
	}
	return nil
}

// CreateGrpcConnection creates a grpc connection for target.
// arg: target ipv4:address[:port]
func CreateGrpcConnection(ctx context.Context, target string) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	opts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	return grpc.DialContext(ctx, target, opts...)
}

func extractError(md *md.MD) (code int32, msg string) {
	c := md.Get("code")
	m := md.Get("msg")
	if len(c) > 0 {
		codeStr, _ := strconv.Atoi(c[0])
		code = int32(codeStr)
	}
	if len(m) > 0 {
		msg = m[0]
	}
	return
}
