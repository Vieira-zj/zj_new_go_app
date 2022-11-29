package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"demo.grpc/grpc.app/pb/account"

	"google.golang.org/grpc"
)

const (
	address = "localhost"
	port    = "50051"
)

func main() {
	callDepositByPb()
	log.Println("protoc client done")
}

//
// deposit server: grpc.reflect/svc_bin/grpc_deposit
//

func callDepositByPb() {
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

func callDepositByProto() {}
