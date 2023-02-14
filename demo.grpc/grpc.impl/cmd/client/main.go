package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"demo.grpc/grpc.impl/pb/account"
	"demo.grpc/grpc.impl/pkg/application"

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
	time.Sleep(time.Second)
	callDepositByInvoke()
	time.Sleep(time.Second)

	callDepositByProto()
	time.Sleep(time.Second)
	callCreateAccountByProto()
	time.Sleep(time.Second)
	callSayHelloByProto()
	log.Println("grpc client done")
}

func callDepositByPb() {
	log.Println(strings.Repeat("*", 10), "call grpc api [deposit] by pb")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// when use grpc.WithBlock() without context here, if grpc server not start or incorrect port, it cuases block.
	opts := []grpc.DialOption{
		grpc.WithInsecure(),
		grpc.WithBlock(),
	}
	conn, err := grpc.DialContext(ctx, fmt.Sprintf("%s:%s", address, port), opts...)
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
	log.Println(strings.Repeat("*", 10), "call grpc api [deposit] by invoke and pb")
	req := &account.DepositRequest{
		Amount: float32(7.1),
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
	method := "/account.DepositService/Deposit"
	body := `{"amount": 10.03}`
	coder := application.GetProtoCoder()
	req, err := coder.BuildReqProtoMessage(method, body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("deposit req:", req.String())

	resp, err := coder.NewRespProtoMessage(method)
	if err != nil {
		log.Fatal(err)
	}

	target := fmt.Sprintf("%s:%s", address, port)
	if err = application.GrpcCall(context.Background(), target, method, req, resp); err != nil {
		log.Fatal(err)
	}
	log.Println("deposit resp:", resp.String())
}

func callCreateAccountByProto() {
	log.Println(strings.Repeat("*", 10), "call grpc api [createAccount] by invoke and proto")
	method := "/account.DepositService/CreateAccount"
	body := `{"account_no": "000011"}`
	coder := application.GetProtoCoder()
	req, err := coder.BuildReqProtoMessage(method, body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("createAccount req:", req.String())

	resp, err := coder.NewRespProtoMessage(method)
	if err != nil {
		log.Fatal(err)
	}

	target := fmt.Sprintf("%s:%s", address, port)
	if err = application.GrpcCall(context.Background(), target, method, req, resp); err != nil {
		log.Fatal(err)
	}
	log.Println("createAccount resp:", resp.String())
}

func callSayHelloByProto() {
	log.Println(strings.Repeat("*", 10), "call grpc api [sayhello] by invoke and proto")
	method := "/greeter.Greeter/SayHello"
	body := `{"name": "foo"}`
	coder := application.GetProtoCoder()
	req, err := coder.BuildReqProtoMessage(method, body)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("sayhello req:", req.String())

	resp, err := coder.NewRespProtoMessage(method)
	if err != nil {
		log.Fatal(err)
	}

	target := fmt.Sprintf("%s:%s", address, port)
	if err = application.GrpcCall(context.Background(), target, method, req, resp); err != nil {
		log.Fatal(err)
	}
	log.Println("sayhello resp:", resp.String())
}
