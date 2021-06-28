package pkg

import (
	"bufio"
	"math/rand"
	"net"
	"strconv"
	"testing"
	"time"
)

func TestEchoHandler(t *testing.T) {
	addr := "127.0.0.1:8080"
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 10; i++ {
		val := strconv.Itoa(rand.Int())
		_, err := conn.Write([]byte(val + "\n"))
		if err != nil {
			t.Fatal(err)
		}

		bufReader := bufio.NewReader(conn)
		line, _, err := bufReader.ReadLine()
		if err != nil {
			t.Fatal(err)
		}

		if string(line) != val {
			t.Fatal("get wrong resopnse")
		}
	}
	conn.Close()

	for i := 0; i < 5; i++ {
		// create idle connection
		if _, err := net.Dial("tcp", addr); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Duration(200) * time.Millisecond)
	}
	time.Sleep(time.Second)
}
