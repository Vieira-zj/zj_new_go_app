package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var client *websocket.Upgrader

func init() {
	client = &websocket.Upgrader{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: time.Duration(3) * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
}

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

// GetDeltaJobResults .
func GetDeltaJobResults(w http.ResponseWriter, r *http.Request) {
	conn, err := client.Upgrade(w, r, nil)
	if err != nil {
		http.Error(w, fmt.Sprintf("create websocket error: %v", err), http.StatusInternalServerError)
	}
	defer func() {
		fmt.Println("close websocket connection")
		conn.Close()
	}()

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			http.Error(w, fmt.Sprintln("receive message error:", err), http.StatusInternalServerError)
			return
		}

		if string(msg) == "sync" {
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(60)*time.Second)
			defer cancel()

			for result := range getMockDeltaJobsResults(ctx) {
				b, err := json.Marshal(result)
				if err != nil {
					http.Error(w, fmt.Sprintln("json marshal error:", err), http.StatusInternalServerError)
				}
				if err := conn.WriteMessage(msgType, b); err != nil {
					http.Error(w, fmt.Sprintln("write message error:", err), http.StatusInternalServerError)
					return
				}
			}
			return
		}
		fmt.Printf("receive message: %s\n", msg)
	}
}
