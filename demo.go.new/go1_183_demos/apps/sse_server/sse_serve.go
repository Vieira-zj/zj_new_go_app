package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type SseHandler struct {
	clients map[chan string]struct{}
}

func NewSseHandler() SseHandler {
	return SseHandler{
		clients: make(map[chan string]struct{}),
	}
}

func (h SseHandler) Serve(c *gin.Context) {
	w, r := c.Writer, c.Request

	// set sse http headers
	w.Header().Add("Content-Type", "text/event-stream")
	w.Header().Add("Cache-Control", "no-cache")
	w.Header().Add("Connection", "keep-alive")

	clientChan := make(chan string)
	h.clients[clientChan] = struct{}{}

	defer func() {
		delete(h.clients, clientChan)
		close(clientChan)
	}()

	for {
		select {
		case msg := <-clientChan:
			fmt.Fprintf(w, "data: %s\n\n", msg)
			w.(http.Flusher).Flush()
		case <-r.Context().Done():
			log.Printf("sse handler exit: %v", r.Context().Err())
			return
		}
	}
}

func (h SseHandler) SimulateEvents() {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// broadcast message
		message := fmt.Sprintf("Server Time: %s", time.Now().Format(time.DateTime))
		for clientCh := range h.clients {
			select {
			case clientCh <- message:
			default:
				// skip when block
			}
		}
	}
}

func SseServe() {
	r := gin.Default()

	r.StaticFile("/", "./static/index.html")

	sseHandler := NewSseHandler()

	r.GET("/healthz", func(c *gin.Context) {
		c.Writer.WriteHeader(http.StatusOK)
	})
	r.GET("/stream", sseHandler.Serve)

	go sseHandler.SimulateEvents()

	if err := r.Run(":8081"); err != nil {
		log.Printf("sse server exit: %v", err)
	}
}
