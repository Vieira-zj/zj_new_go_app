package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

const (
	modeServer = "server"
	modeDial   = "dial"
)

var (
	mode string
	port string
	help bool
)

func init() {
	flag.StringVar(&port, "p", ":8081", "Listen or serve port.")
	flag.StringVar(&mode, "m", modeServer, "Run mode, server (default) or dial.")
	flag.BoolVar(&help, "h", false, "Help")
	flag.Parse()
}

func main() {
	if help {
		flag.Usage()
		return
	}

	switch mode {
	case modeServer:
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		if err := serveTCP(ctx, port); err != nil {
			panic(err)
		}
		<-ctx.Done()
		stop()
	case modeDial:
		if err := dialTCP(port); err != nil {
			panic(err)
		}
	default:
		fmt.Println("invalid mode:", mode)
		os.Exit(99)
	}

	os.Exit(0)
}

//
// Serve Tcp
//
// Run: ./tcpserver
// Test: telnet 127.0.0.1 8081
//

var (
	tcpConns []net.Conn
)

func serveTCP(ctx context.Context, port string) error {
	fmt.Printf("tcp serve at: [%s]\n", port)
	listen, err := net.Listen("tcp", port)
	if err != nil {
		return fmt.Errorf("listen tcp error: %v", err)
	}
	defer listen.Close()

	go func() {
		<-ctx.Done()
		fmt.Println("tcp server exit:", ctx.Err())
		for _, conn := range tcpConns {
			conn.Close()
		}
		listen.Close()
	}()

	for {
		conn, err := listen.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				fmt.Println("tcp conn is already closed")
				return nil
			}
			return fmt.Errorf("get tcp conn error: %v", err)
		}
		tcpConns = append(tcpConns, conn)
		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("tcp conn read eof")
				return
			}
			if errors.Is(err, net.ErrClosed) {
				fmt.Println("tcp conn is closed")
				return
			}
			panic(fmt.Sprintf("tcp conn read error: %v", err))
		}

		fmt.Print("read msg:", msg)
		msg = strings.ToLower(msg)
		if strings.HasPrefix(msg, "exit") || strings.HasPrefix(msg, "bye") {
			fmt.Println("tcp conn exit")
			if _, err := conn.Write([]byte("[server] bye\n")); err != nil {
				panic(fmt.Sprintf("tcp conn write error: %v", err))
			}
			return
		}
		if _, err := conn.Write([]byte("[server] got msg\n")); err != nil {
			panic(fmt.Sprintf("tcp conn write error: %v", err))
		}
	}
}

//
// Dial Tcp
//
// Run: ./tcpserver -m=dial
//

var (
	input string
	resp  []byte
)

func dialTCP(port string) error {
	fmt.Printf("tcp listen at: [%s]\n", port)
	conn, err := net.Dial("tcp", port)
	if err != nil {
		return fmt.Errorf("dial tcp error: %v", err)
	}
	defer conn.Close()

	for {
		fmt.Print("input: ")
		if _, err := fmt.Scanln(&input); err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("input eof")
				return nil
			}
			return fmt.Errorf("scanln error: %v", err)
		}

		if _, err := conn.Write([]byte(fmt.Sprintln(input))); err != nil {
			return fmt.Errorf("tcp conn write error: %v", err)
		}

		reader := bufio.NewReader(conn)
		resp, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				fmt.Println("tcp conn read eof")
			}
			panic(fmt.Sprintln("tcp conn read error:", err))
		}
		fmt.Printf("[client] got response: %s", resp)

		if input == "exit" || input == "bye" {
			fmt.Println("tcp listen exit")
			return nil
		}
	}
}
