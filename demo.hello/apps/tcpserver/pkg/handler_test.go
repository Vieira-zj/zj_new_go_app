package pkg

import (
	"bufio"
	"io"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestNewLine(t *testing.T) {
	for _, line := range []string{
		"hello",
		"hello\n",
	} {
		t.Log("is new line:", line[len(line)-1] == '\n')
	}
}

func TestTcpEchoHandler(t *testing.T) {
	addr := "127.0.0.1:8080"
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		time.Sleep(time.Second) // wait read
		conn.Close()
	}()

	// loop: read
	go func() {
		bufReader := bufio.NewReader(conn)
		for {
			// block read until receive data; if eof or connect close from remote, raise error immediately
			line, _, err := bufReader.ReadLine()
			if err != nil {
				if err == io.EOF {
					t.Log("connect close")
				} else {
					t.Log(err)
				}
				return
			}
			t.Log("read line: " + string(line))
		}
	}()

	r := rand.New(rand.NewSource(time.Now().UnixMilli()))

	// loop: write
	for i := 0; i < 10; i++ {
		val := strconv.Itoa(r.Intn(100))
		_, err := conn.Write([]byte("int:" + val + "\n"))
		if err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Second)
	}

	t.Log("tcp test done")
}

func TestCreateIdleConn(t *testing.T) {
	addr := "127.0.0.1:8080"
	for i := 0; i < 3; i++ {
		if _, err := net.Dial("tcp", addr); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Duration(300) * time.Millisecond)
	}

	time.Sleep(time.Second)
	t.Log("create connect done")
}
