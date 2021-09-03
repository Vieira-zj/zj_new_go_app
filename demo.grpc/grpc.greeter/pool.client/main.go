package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"demo.grpc/grpc.greeter/proto/message"

	"github.com/silenceper/pool"
	"google.golang.org/grpc"
)

/*
Requester with pool of connections
*/

// Requester manage grpc request.
type Requester struct {
	addr      string
	service   string
	method    string
	timeoutMs uint
	pool      pool.Pool
}

// NewRequester returns an Requester instance.
func NewRequester(addr string, service string, method string, timeoutMs uint, poolsize int) (*Requester, error) {
	factory := func() (interface{}, error) {
		return grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	}

	closef := func(v interface{}) error {
		return v.(*grpc.ClientConn).Close()
	}

	// 创建一个连接池: 初始化3, 最大连接200, 最大空闲10
	// 连接最大空闲时间15s, 超过该时间的连接将会关闭, 可避免空闲时连接EOF, 自动失效的问题
	poolConfig := &pool.Config{
		InitialCap:  3,
		MaxCap:      poolsize,
		MaxIdle:     10,
		IdleTimeout: 15 * time.Second,
		Factory:     factory,
		Close:       closef,
	}

	p, err := pool.NewChannelPool(poolConfig)
	if err != nil {
		log.Println("New a pool failed:", err)
		return nil, err
	}

	return &Requester{
		addr:      addr,
		service:   service,
		method:    method,
		timeoutMs: timeoutMs,
		pool:      p,
	}, nil
}

func (r *Requester) getRealMethodName() string {
	return fmt.Sprintf("/%s/%s", r.service, r.method)
}

// Call invokes grpc service.
func (r *Requester) Call(req interface{}, resp interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.timeoutMs)*time.Millisecond)
	defer cancel()

	conn, err := r.pool.Get()
	if err != nil {
		log.Println("Get grpc client connection failed.")
		return err
	}
	defer r.pool.Put(conn)

	if err := conn.(*grpc.ClientConn).Invoke(ctx, r.getRealMethodName(), req, resp); err != nil {
		fmt.Fprintf(os.Stderr, err.Error())
		return err
	}
	return nil
}

/*
Main

Refer: https://github.com/bugVanisher/grequester
*/

var (
	addr      = "localhost:50051"
	service   = "proto.Greeter"
	method    = "SayHello"
	timeoutMs = uint(800)
	poolsize  = 15
)

func main() {
	requester, err := NewRequester(addr, service, method, timeoutMs, poolsize)
	if err != nil {
		panic(err)
	}

	for _, name := range []string{"zhengjin", "vieira", "henry"} {
		req := &message.HelloRequest{}
		reqJSONText := fmt.Sprintf(`{"name":"%s"}`, name)
		if err := json.Unmarshal([]byte(reqJSONText), req); err != nil {
			panic(err)
		}
		log.Printf("Send: %s", reqJSONText)

		resp := &message.HelloReply{}
		if err := requester.Call(req, resp); err != nil {
			panic(err)
		}

		respJSON, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}
		log.Println("Reply:", string(respJSON))
		time.Sleep(time.Duration(500) * time.Millisecond)
	}

	log.Println("grpc demo done.")
}
