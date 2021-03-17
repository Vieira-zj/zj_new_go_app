package account

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"

	pb "demo.grpc/grpc.unittest/proto/account"
)

func dialer() func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(1024 * 1024)

	server := grpc.NewServer()
	pb.RegisterDepositServiceServer(server, &DepositServer{})

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatal(err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}

func TestDepositServerDeposit(t *testing.T) {
	tests := []struct {
		name    string
		amount  float32
		res     *pb.DepositResponse
		errCode codes.Code
		errMsg  string
	}{
		{
			"invalid request with negative amount",
			-1.11,
			nil,
			codes.InvalidArgument,
			fmt.Sprintf("cannot deposit %v", -1.11),
		},
		{
			"valid request with non negative amount",
			0.00,
			&pb.DepositResponse{Ok: true},
			codes.OK,
			"",
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(ctx, "", grpc.WithInsecure(), grpc.WithContextDialer(dialer()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	client := pb.NewDepositServiceClient(conn)
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := &pb.DepositRequest{Amount: test.amount}
			resp, err := client.Deposit(ctx, request)
			if err != nil {
				if er, ok := status.FromError(err); ok {
					if er.Code() != test.errCode {
						t.Error("error code: expected", codes.InvalidArgument, "received", er.Code())
					}
					if er.Message() != test.errMsg {
						t.Error("error message: expected", test.errMsg, "received", er.Message())
					}
				}
			}

			if resp != nil {
				if resp.GetOk() != test.res.GetOk() {
					t.Error("response: expected", test.res.GetOk(), "received", resp.GetOk())
				}
			}
		})
	}
}
