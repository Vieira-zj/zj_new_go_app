package main

import (
	"fmt"
	"os"

	"demo.hello/apps/pbparser/pkg"
	"github.com/emicklei/proto"
)

var visit pkg.TestVisitor

func handleService(s *proto.Service) {
	fmt.Println("service:", s.Name)
	for _, ele := range s.Elements {
		ele.Accept(visit)
	}
}

func handleMessage(m *proto.Message) {
	fmt.Println("message:", m.Name)
	for _, ele := range m.Elements {
		ele.Accept(visit)
	}
}

func main() {
	// refer: https://github.com/emicklei/proto

	pbFilePath := "/Users/jinzheng/Workspaces/zj_repos/zj_go2_project/demo.grpc/gateway/proto/demo/hello/service2.proto"
	reader, err := os.Open(pbFilePath)
	if err != nil {
		panic(err)
	}
	defer reader.Close()

	parser := proto.NewParser(reader)
	definition, err := parser.Parse()
	if err != nil {
		panic(err)
	}

	proto.Walk(definition, proto.WithService(handleService), proto.WithMessage(handleMessage))
	fmt.Println("pb parse demo done.")
}
