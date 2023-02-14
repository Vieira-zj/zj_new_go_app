package interceptor

import (
	"context"
	"log"
	"time"

	"demo.grpc/grpc.app/pkg/codec"
	"google.golang.org/grpc"
)

func LoggingClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		log.Println("logging client interceptor start")
		defer func() {
			log.Println("logging client interceptor end")
		}()

		log.Println("grpc call method:", method)
		protoJson := codec.ProtoJson{}
		b, err := protoJson.Marshal(req)
		if err != nil {
			log.Println("protoJson marshal error:", err)
		}
		log.Println("grpc call request:", string(b))

		if err = invoker(ctx, method, req, reply, cc, opts...); err != nil {
			return err
		}

		b, err = protoJson.Marshal(reply)
		if err != nil {
			log.Println("protoJson marshal error:", err)
		}
		log.Println("grpc call reply:", string(b))
		return nil
	}
}

func RetryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		log.Println("retry client interceptor start")
		defer func() {
			log.Println("retry client interceptor end")
		}()

		t := time.NewTicker(time.Second)
		defer t.Stop()

		var err error
	outer:
		for i := 0; i < 3; i++ {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-t.C:
				if err = invoker(ctx, method, req, reply, cc, opts...); err != nil {
					log.Println("grpc error:", err)
					log.Println("grpc invoke retry:", i+1)
				} else {
					break outer
				}
			}
		}
		return err
	}
}
