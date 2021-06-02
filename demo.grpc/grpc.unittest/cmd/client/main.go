package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"demo.grpc/grpc.unittest/internal/account"
	"google.golang.org/grpc"
)

var (
	port   string
	amount float64
	help   bool
)

func main() {
	flag.StringVar(&port, "p", "50051", "grpc listen port.")
	flag.Float64Var(&amount, "m", 0, "deposit amount.")
	flag.BoolVar(&help, "h", false, "help.")
	flag.Parse()

	if help {
		flag.Usage()
		return
	}

	log.Println("Grpc Client running ...")
	conn, err := grpc.Dial(fmt.Sprintf(":%s", port), grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	resp, err := account.NewDepositClient(conn, time.Second).Deposit(context.Background(), float32(amount))
	if err != nil {
		log.Println(err)
	}
	log.Println("Deposit Results:", resp)
}
