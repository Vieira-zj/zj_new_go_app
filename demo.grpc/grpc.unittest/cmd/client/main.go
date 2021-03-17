package main

import (
	"context"
	"log"
	"time"

	"demo.grpc/grpc.unittest/internal/account"
	"google.golang.org/grpc"
)

func main() {
	log.Println("Grpc Client running ...")

	conn, err := grpc.Dial(":50051", grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	resp, err := account.NewDepositClient(conn, time.Second).Deposit(context.Background(), 1990.01)
	if err != nil {
		log.Println(err)
	}
	log.Println("Deposit Results:", resp)
}
