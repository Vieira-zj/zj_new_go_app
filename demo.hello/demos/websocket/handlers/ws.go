package handlers

import (
	"fmt"
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"time"

	"demo.hello/utils"
	"github.com/gorilla/websocket"
)

const connsLimit = 4

var (
	client     *websocket.Upgrader
	connsCount int
)

func init() {
	client = &websocket.Upgrader{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: 3 * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}

// Echo .
func Echo(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("[wsEcho]: goroutines count: %d\n", runtime.NumGoroutine())
	connsCount++
	if connsCount > connsLimit {
		fmt.Printf("[wsEcho]: exceed max ws connections number: %d\n", connsLimit)
		return
	}

	conn, err := client.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("[wsEcho]: create websocket error: %v\n", err)
		return
	}
	defer func() {
		fmt.Println("[wsEcho]: close ws connection")
		conn.Close()
	}()

	cbName := strconv.Itoa(time.Now().Nanosecond())
	cb := utils.Callback{
		Name: cbName,
		Fn: func(args ...interface{}) {
			msg := args[0].(string)
			if err = conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				fmt.Printf("[callback:%s]: write message error: %v\n", cbName, err)
			}
		},
	}
	eventBusSvr.Register(echoChannel, cb)
	defer func() {
		fmt.Println(fmt.Println("[wsEcho]: unregister callback:", cbName))
		eventBusSvr.Unregister(echoChannel, cb)
	}()

	for {
		// if no message, blocked here
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("[wsEcho]: receive message error:", err)
			return
		}
		fmt.Printf("[wsEcho]: receive message [%s] from [%s]\n", msg, conn.RemoteAddr().String())

		if strings.ToLower(string(msg)) == "exit" {
			fmt.Println("[wsEcho]: exit")
			connsCount--
			return
		}
	}
}
