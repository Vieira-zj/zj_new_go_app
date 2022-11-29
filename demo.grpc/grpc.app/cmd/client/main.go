package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"demo.grpc/grpc.app/pb/account"
	"demo.grpc/grpc.app/pkg/application"
	"demo.grpc/grpc.app/pkg/protoc"
	"github.com/jhump/protoreflect/desc"

	"google.golang.org/grpc"
)

const (
	address = "localhost"
	port    = "50051"
)

func init() {
	subPath := "Workspaces/zj_repos/zj_new_go_project/demo.grpc"
	os.Setenv("PROJECT_ROOT", filepath.Join(os.Getenv("HOME"), subPath))
}

// deposit grpc server: grpc.reflect/svc_bin/grpc_deposit

func main() {
	callDepositByPb()
	callDepositByInvoke()
	time.Sleep(time.Second)
	callDepositByProto()
	time.Sleep(time.Second)
	log.Println("grpc client done")
}

func callDepositByPb() {
	log.Println("call grpc api [deposit] by pb")
	// NOTE: do not use grpc.WithBlock() here, if grpc server not start or incorrect port, it cuases block.
	conn, err := grpc.Dial(fmt.Sprintf("%s:%s", address, port), grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Did not connect: %v", err)
	}
	defer conn.Close()

	for i := 0; i < 3; i++ {
		i := i
		func() {
			log.Println("grpc invoke to deposit")
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()

			client := account.NewDepositServiceClient(conn)
			resp, err := client.Deposit(ctx, &account.DepositRequest{
				Amount: float32(i),
			})
			if err != nil {
				log.Fatalln("deposit error:", err)
			}
			log.Println("deposit resp:", resp.GetOk())
		}()
		time.Sleep(time.Second)
	}
}

func callDepositByInvoke() {
	log.Println("call grpc api [deposit] by invoke and pb")
	req := &account.DepositRequest{
		Amount: float32(7),
	}
	resp := &account.DepositResponse{}

	target := fmt.Sprintf("%s:%s", address, port)
	fullMethod := "/account.DepositService/Deposit"
	if err := application.GrpcCall(context.Background(), target, fullMethod, req, resp); err != nil {
		log.Fatal(err)
	}
	log.Println("deposit resp:", resp.GetOk())
}

func callDepositByProto() {
	// process: load/parse proto => coder => req,resp (proto message) => grpc invoke
	log.Println("call grpc api [deposit] by invoke and proto")
	mds, err := loadProto()
	if err != nil {
		log.Fatal(err)
	}
	for k, md := range mds {
		log.Println("load:", k, md.GetName())
	}

	method := "/account.DepositService/Deposit"
	body := `{"amount": 10}`
	coder := protoc.NewCoder(mds)
	req, err := coder.BuildReqProtoMessage(method, body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("req:", req.String())

	resp, err := coder.NewRespProtoMessage(method)
	if err != nil {
		log.Fatal(err)
	}

	target := fmt.Sprintf("%s:%s", address, port)
	if err = application.GrpcCall(context.Background(), target, method, req, resp); err != nil {
		log.Fatal(err)
	}
	log.Println("resp:", resp.String())
}

func loadProto() (map[string]*desc.MethodDescriptor, error) {
	path := filepath.Join(os.Getenv("PROJECT_ROOT"), "grpc.app/proto")
	dirPaths, err := protoc.GetAllProtoDirs(path)
	if err != nil {
		return nil, err
	}
	return protoc.LoadProtoFiles(dirPaths...)
}
