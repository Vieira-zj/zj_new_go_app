package account

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	pb "demo.grpc/grpc.unittest/proto/account"
)

// DepositClient deposit client impl.
type DepositClient struct {
	conn    *grpc.ClientConn
	timeout time.Duration
}

// NewDepositClient returns a DepositClient instance.
func NewDepositClient(conn *grpc.ClientConn, timeout time.Duration) *DepositClient {
	return &DepositClient{
		conn:    conn,
		timeout: timeout,
	}
}

// Deposit deposit api impl.
func (d *DepositClient) Deposit(ctx context.Context, amount float32) (bool, error) {
	client := pb.NewDepositServiceClient(d.conn)

	request := &pb.DepositRequest{Amount: amount}

	ctx, cancel := context.WithTimeout(ctx, d.timeout)
	defer cancel()

	resp, err := client.Deposit(ctx, request)
	if err != nil {
		if er, ok := status.FromError(err); ok {
			return false, fmt.Errorf("grpc: %s, %s", er.Code(), er.Message())
		}
		return false, fmt.Errorf("server: %s", err.Error())
	}
	return resp.GetOk(), nil
}
