package main

import (
	"context"
	"log"
	"sync"
	"time"

	pb "demo.grpc/grpc.greeter/proto"
	"demo.grpc/grpc.greeter/proto/message"
	"google.golang.org/grpc"
)

/*
grpc client 使用单例模式，支持 长/短 连接切换
*/

// inernalGreeterClient 单例（一个grpc连接）
var inernalGreeterClientInstance pb.GreeterClient

// 互斥锁，对 inernalGreeterClientInstance 提供并发访问保护
var inernalGreeterClientMutex sync.Mutex

type inernalGreeterClient struct {
	address     string
	dialOptions []grpc.DialOption
}

func (i *inernalGreeterClient) SayHello(ctx context.Context, in *message.HelloRequest, opts ...grpc.CallOption) (*message.HelloReply, error) {
	useLongConnection := len(opts) == 0

	// 如果启⽤了⻓连接, 且 client 已被初始化, 直接进⾏⽅法调⽤
	if useLongConnection && inernalGreeterClientInstance != nil {
		log.Println("re-use long connection for inernalGreeterClient")
		return inernalGreeterClientInstance.SayHello(ctx, in, opts...)
	}

	// client 初始化
	c, conn, err := getGreeterClient(i.address, i.dialOptions...)
	if err != nil {
		return nil, err
	}

	if useLongConnection {
		inernalGreeterClientMutex.Lock()
		defer inernalGreeterClientMutex.Unlock()
		// DCL 双重检查, 确保实例只会被初始化⼀次
		if inernalGreeterClientInstance == nil {
			inernalGreeterClientInstance = c
			log.Println("long connection established for inernalGreeterClient")
		} else {
			// 当未通过双重检查时, 关闭当前连接, 避免连接泄露
			defer conn.Close()
			log.Println("long connection for inernalGreeterClient has been established, going to close current connection")
		}
	} else {
		// 为短连接时, 在⽅法调⽤后关闭连接
		defer conn.Close()
	}
	return inernalGreeterClientInstance.SayHello(ctx, in, opts...)
}

func getGreeterClient(address string, opts ...grpc.DialOption) (pb.GreeterClient, *grpc.ClientConn, error) {
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, nil, err
	}
	c := pb.NewGreeterClient(conn)
	return c, conn, nil
}

/*
Main
*/

func main() {
	address := "localhost:50051"
	client := &inernalGreeterClient{
		address: address,
		dialOptions: []grpc.DialOption{
			grpc.WithInsecure(), grpc.WithBlock(),
		},
	}

	names := []string{"zhengjin", "zj", "vieira", "henry", "zhengj"}
	for _, name := range names {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		r, err := client.SayHello(ctx, &message.HelloRequest{Name: name})
		if err != nil {
			log.Fatalf("Could not greet: %v", err)
		}
		log.Printf("Greeting: %s", r.GetMessage())
	}
}
