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
		http.Error(w, fmt.Sprintf("create websocket error: %v", err), http.StatusInternalServerError)
	}
	defer func() {
		fmt.Println("close websocket connection")
		conn.Close()
	}()

	for {
		// if no message, blocked here
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			http.Error(w, fmt.Sprintln("receive message error:", err), http.StatusInternalServerError)
			return
		}
		fmt.Printf("receive message [%s] from [%s]\n", msg, conn.RemoteAddr().String())

		retMsg := fmt.Sprintf("hello %s", msg)
		if string(msg) == "exit" {
			retMsg = "bye"
		}
		if err := conn.WriteMessage(msgType, []byte(retMsg)); err != nil {
			http.Error(w, fmt.Sprintln("write message error:", err), http.StatusInternalServerError)
			return
		} else if err == nil && string(msg) == "exit" {
			return
		}
	}
}

// SyncDeltaJobResults .
func SyncDeltaJobResults(w http.ResponseWriter, r *http.Request) {
	conn, err := client.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("create websocket error: %v", err), http.StatusInternalServerError)
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
				http.Error(w, fmt.Sprintln("json marshal error:", err), http.StatusInternalServerError)
			}
			if err = conn.WriteMessage(websocket.TextMessage, b); err != nil {
				http.Error(w, fmt.Sprintln("write message error:", err), http.StatusInternalServerError)
				return
			}
		},
	}
	EventBus.Register(channel, callback)
	defer EventBus.Unregister(channel, callback)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			http.Error(w, fmt.Sprintln("receive message error:", err), http.StatusInternalServerError)
			return
		}
		fmt.Printf("receive message: %s\n", msg)

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
		defer cancel()
		if err = mock.getDeltaJobResults(ctx); err != nil {
			http.Error(w, fmt.Sprintln("get delta job results error:", err), http.StatusInternalServerError)
		}
	}
}
