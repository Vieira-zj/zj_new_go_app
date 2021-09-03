package helper

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
)

func TestAllMethodsViaReflection(t *testing.T) {
	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	dial := func(ctx context.Context, addr string) (net.Conn, error) {
		dialer := &net.Dialer{}
		return dialer.DialContext(ctx, "tcp", addr)
	}
	opts = append(opts, grpc.WithContextDialer(dial))

	dialTime := time.Duration(5) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), dialTime)
	defer cancel()

	addr := "localhost:50051"
	cc, err := grpc.DialContext(ctx, addr, opts...)
	if err != nil {
		t.Fatal(err)
	}

	ret, err := AllMethodsViaReflection(ctx, cc)
	for svc, methods := range ret {
		fmt.Println("service:", svc)
		for _, name := range methods {
			fmt.Println("\trpc method:", name.GetFullyQualifiedName())
		}
	}
}
