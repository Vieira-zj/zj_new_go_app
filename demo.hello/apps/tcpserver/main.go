package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
)

// ListenAndServe .
func ListenAndServe(address string) {
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatal(fmt.Sprintf("tcp listen error: %v", err))
	}
	defer listener.Close()
	log.Println(fmt.Sprintf("tcp bind: %s, start listening...", address))

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(fmt.Sprintf("accept err: %v", err))
		}
		go Handler(conn)

	}
}

// Handler .
func Handler(conn net.Conn) {
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("connection close")
			} else {
				log.Println(err)
			}
			return
		}
		conn.Write([]byte(msg))
	}
}

func main() {
	ListenAndServe(":8080")
}
