package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"demo.hello/utils"
	"github.com/gorilla/websocket"
)

// EchoMessage .
func EchoMessage(w http.ResponseWriter, r *http.Request) {
	conn, err := client.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("create websocket error: %v", err)
		return
	}
	defer func() {
		fmt.Println("close websocket connection")
		conn.Close()
	}()

	for {
		// if no message, blocked here
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("receive message error:", err)
			return
		}
		fmt.Printf("receive message [%s] from [%s]\n", msg, conn.RemoteAddr().String())

		retMsg := fmt.Sprintf("hello %s", msg)
		if string(msg) == "exit" {
			fmt.Println("ws echo exit")
			return
		}
		err = conn.WriteMessage(msgType, []byte(retMsg))
		if err != nil {
			fmt.Println("write message error:", err)
			return
		}
	}
}

// SyncDeltaJobResults .
func SyncDeltaJobResults(w http.ResponseWriter, r *http.Request) {
	conn, err := client.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("create websocket error: %v", err)
		return
	}
	defer func() {
		fmt.Println("close websocket connection")
		conn.Close()
	}()

	connKey := strconv.Itoa(time.Now().Nanosecond())
	callback := utils.Callback{
		Name: connKey,
		Fn: func(result ...interface{}) {
			b, err := json.Marshal(result[0])
			if err != nil {
				b = []byte(fmt.Sprintln("json marshal error:", err))
			}
			if err = conn.WriteMessage(websocket.TextMessage, b); err != nil {
				fmt.Println("write message error:", err)
				return
			}
		},
	}

	EventBus.Register(channel, callback)
	defer EventBus.Unregister(channel, callback)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("receive message error:", err)
			return
		}
		fmt.Printf("receive message: %s\n", msg)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
		defer cancel()
		if err = mock.getDeltaJobResults(ctx); err != nil {
			fmt.Println("get delta job results error:", err)
		}
	}
}
