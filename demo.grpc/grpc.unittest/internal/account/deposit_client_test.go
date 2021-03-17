package account

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	pb "demo.grpc/grpc.unittest/proto/account"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
)

type mockDepositServer struct {
	pb.UnimplementedDepositServiceServer
}

func (*mockDepositServer) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.DepositResponse, error) {
	if req.GetAmount() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "cannot deposit %v", req.GetAmount())
	}
	return &pb.DepositResponse{Ok: true}, nil
}

func dialerClient() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()

	pb.RegisterDepositServiceServer(server, &mockDepositServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestDepositClientDeposit(t *testing.T) {
	tests := []struct {
		name   string
		amount float32
		res    bool
		err    error
	}{
		{
			"invalid request with negative amount",
			-1.11,
			false,
			fmt.Errorf("grpc: InvalidArgument, cannot deposit %v", -1.11),
		},
		{
			"valid request with non negative amount",
			0.00,
			true,
			nil,
		},
	}

	conn, err := grpc.DialContext(context.Background(), "", grpc.WithInsecure(), grpc.WithContextDialer(dialerClient()))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			resp, err := NewDepositClient(conn, time.Second).Deposit(context.Background(), test.amount)
			if err != nil && errors.Is(err, test.err) {
				t.Error("error: expected", test.err, "received", err)
			}

			if resp != test.res {
				t.Error("error: expected", test.res, "received", resp)
			}
		})
	}
}
