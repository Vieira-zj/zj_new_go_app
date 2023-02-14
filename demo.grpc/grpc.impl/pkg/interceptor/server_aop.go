package interceptor

import (
	"context"
	"log"
	"runtime/debug"

	"demo.grpc/grpc.app/pkg/codec"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamServerInterceptor

func LoggingStreamServerInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		log.Println("logging stream server interceptor")
		log.Printf("info: full_method=%s, is_client_stream=%v, is_server_stream=%v",
			info.FullMethod, info.IsClientStream, info.IsServerStream)
		return handler(srv, ss)
	}
}

func RecoverStreamServerInterceptor() grpc.StreamServerInterceptor {
	log.Println("add recover stream server interceptor")
	opt := grpc_recovery.WithRecoveryHandlerContext(RecoveryHandlerFunc())
	return grpc_recovery.StreamServerInterceptor(opt)
}

func RecoveryHandlerFunc() grpc_recovery.RecoveryHandlerFuncContext {
	return func(ctx context.Context, p interface{}) error {
		log.Printf("panic=%v, stack=%s", p, debug.Stack())
		return status.Error(codes.Internal, "server error")
	}
}

// ServerInterceptor

func LoggingServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		log.Println("logging server interceptor")
		log.Println("invoke method:", info.FullMethod)
		resp, err := handler(ctx, req)
		if err != nil {
			return nil, err
		}

		protoJson := codec.ProtoJson{}
		b, err := protoJson.Marshal(resp)
		if err != nil {
			log.Println("protoJson marshal error:", err)
		}
		log.Println("server resp:", string(b))

		return resp, nil
	}
}
