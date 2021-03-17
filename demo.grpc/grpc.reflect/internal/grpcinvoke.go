package internal

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/fullstorydev/grpcurl"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/jhump/protoreflect/desc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type rpcMetadata struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type rpcInput struct {
	TimeoutSeconds float32           `json:"timeout_seconds"`
	Metadata       []rpcMetadata     `json:"metadata"`
	Data           []json.RawMessage `json:"data"`
}

type rpcResult struct{}

func (*rpcResult) OnResolveMethod(*desc.MethodDescriptor) {}

func (*rpcResult) OnSendHeaders(metadata.MD) {}

func (r *rpcResult) OnReceiveHeaders(md metadata.MD) {
	fmt.Println("OnReceiveHeaders:")
	for k, v := range md {
		fmt.Printf("%s=%v", k, v)
	}
	fmt.Println()
}

func (r *rpcResult) OnReceiveResponse(m proto.Message) {
	fmt.Println("OnReceiveResponse:", m.String())
}

func (r *rpcResult) OnReceiveTrailers(stat *status.Status, md metadata.MD) {
	fmt.Println("OnReceiveTrailers:", stat.Code(), stat.Message())
}

// CallGrpc invokes rpc api.
func CallGrpc(ctx context.Context, descSource grpcurl.DescriptorSource, cc *grpc.ClientConn, method string, body string) error {
	var input rpcInput
	if err := json.Unmarshal([]byte(body), &input); err != nil {
		return err
	}

	requestFunc := func(msg proto.Message) error {
		if len(input.Data) == 0 {
			return io.EOF
		}
		req := input.Data[0]
		input.Data = input.Data[1:]
		if err := jsonpb.Unmarshal(bytes.NewReader([]byte(req)), msg); err != nil {
			return status.Errorf(codes.InvalidArgument, err.Error())
		}
		return nil
	}

	hdrs := make([]string, len(input.Metadata))
	for i, hdr := range input.Metadata {
		hdrs[i] = fmt.Sprintf("%s: %s", hdr.Name, hdr.Value)
	}

	if input.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		timeout := time.Duration(input.TimeoutSeconds) * time.Second
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	if err := grpcurl.InvokeRPC(ctx, descSource, cc, method, hdrs, &rpcResult{}, requestFunc); err != nil {
		return err
	}
	return nil
}
