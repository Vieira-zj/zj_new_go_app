package account

import (
	"context"
	"log"

	pb "demo.grpc/grpc.unittest/proto/account"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// DepositServer deposit server impl.
type DepositServer struct {
	pb.UnimplementedDepositServiceServer
}

// Deposit deposit service impl.
func (*DepositServer) Deposit(ctx context.Context, req *pb.DepositRequest) (*pb.DepositResponse, error) {
	log.Printf("Get Amount: %.2f", req.GetAmount())
	if req.GetAmount() < 0 {
		return nil, status.Errorf(codes.InvalidArgument, "cannot deposit %v", req.GetAmount())
	}
	return &pb.DepositResponse{Ok: true}, nil
}
