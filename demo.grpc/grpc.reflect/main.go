package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strings"
	"time"

	"demo.grpc/grpc.reflect/internal"
	"github.com/fullstorydev/grpcurl"
	"github.com/jhump/protoreflect/grpcreflect"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

var debug bool

func runGrpcApp(target, method, body string) error {
	dialTime := time.Duration(10) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), dialTime)
	defer cancel()

	cc, err := grpcurl.BlockingDial(ctx, "tcp", target, nil)
	if err != nil {
		return err
	}

	// get src grpc api meta desc info by reflection
	md := grpcurl.MetadataFromHeaders(nil)
	refCtx := metadata.NewOutgoingContext(ctx, md)
	refClient := grpcreflect.NewClient(refCtx, reflectpb.NewServerReflectionClient(cc))
	descSource := grpcurl.DescriptorSourceFromServer(ctx, refClient)

	if debug {
		fmt.Println(strings.Repeat("*", 30), "proto info")
		if err := internal.PrintProtoFileInfo(descSource); err != nil {
			return err
		}
		fmt.Println()

		fmt.Println(strings.Repeat("*", 30), "service info")
		if err := internal.PrintGrpcServiceInfo(descSource); err != nil {
			return err
		}
		fmt.Println()
	}

	if len(method) > 0 {
		validate, err := internal.IsMethodValidate(descSource, method)
		if err != nil {
			return err
		}
		if validate {
			fmt.Println(strings.Repeat("*", 30), "rpc invoke")
			if err := internal.CallGrpc(ctx, descSource, cc, method, body); err != nil {
				return err
			}
		}
	}
	return nil
}

func main() {
	// pre-condition: reflection of grpc service is enabled.
	target := flag.String("addr", "", "Target address of grpc service.")
	method := flag.String("method", "", "Grpc method to be invoked.")
	body := flag.String("body", `{"metadata":[],"data":[{"name":"tester"}]}`, "Grpc request body.")

	flag.BoolVar(&debug, "debug", false, "Print grpc service meta info.")
	help := flag.Bool("help", false, "Help.")

	flag.Parse()
	if *help {
		flag.Usage()
		return
	}

	if len(*target) <= 0 {
		panic(errors.New("Target address of grpc service is empty"))
	}

	if err := runGrpcApp(*target, *method, *body); err != nil {
		panic(err)
	}
	fmt.Println("grpc reflect tool Done.")
}
