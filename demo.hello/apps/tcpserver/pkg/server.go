package pkg

import (
	"context"
	"fmt"
	"log"
	"net"
)

// ListenAndServe .
func ListenAndServe(address string, handler TCPHandler, closeCh chan struct{}) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(fmt.Sprintf("tcp listen error: %v", err))
	}
	defer func() {
		listener.Close() // listener.Accept() will return err immediately
		handler.Close()
	}()
	log.Println(fmt.Sprintf("tcp bind: %s, start listening ...", address))

	go func() {
		<-closeCh
		log.Println("tcp server shutdown ...")
		listener.Close()
		handler.Close()
	}()

	for {
		// 同时处理多个请求（每个请求对应一个连接）
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(fmt.Sprintf("accept err: %v", err))
		}
		go handler.Handle(context.Background(), conn)
		log.Println(fmt.Sprintf("tcp conns counnt: %d", handler.ConnsCount()))
	}
}
