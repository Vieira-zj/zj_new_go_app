package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"demo.grpc/grpc.impl/pkg/application"
)

func init() {
	subPath := "Workspaces/zj_repos/zj_new_go_project/demo.grpc"
	os.Setenv("PROJECT_ROOT", filepath.Join(os.Getenv("HOME"), subPath))
	application.GetProtoCoder()
}

func main() {
	port := "50051"
	if err := application.RunGrpcServer(port); err != nil {
		log.Fatal(err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	log.Println("grpc server exit")
}
