package main

import (
	"fmt"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// refer: https://github.com/gin-gonic/examples/blob/master/server-sent-event/main.go

type ClientChan chan string

const ClientChanCtxKey = "ClientChan"

type Event struct {
	MessageChan   chan string
	Clients       map[ClientChan]struct{}
	NewClients    chan ClientChan
	ClosedClients chan ClientChan
}

func NewStreamServer() (event *Event) {
	event = &Event{
		MessageChan:   make(chan string),
		Clients:       make(map[ClientChan]struct{}),
		NewClients:    make(chan ClientChan),
		ClosedClients: make(chan ClientChan),
	}

	go event.Listen()
	return
}

func (stream Event) Listen() {
	for {
		select {
		// add new available client
		case client := <-stream.NewClients:
			stream.Clients[client] = struct{}{}
			log.Printf("client added: %d registered clients", len(stream.Clients))

		// remove closed client
		case client := <-stream.ClosedClients:
			delete(stream.Clients, client)
			close(client)
			log.Printf("removed client: %d registered clients", len(stream.Clients))

		// broadcast message to client
		case eventMsg := <-stream.MessageChan:
			for clientMsgChan := range stream.Clients {
				clientMsgChan <- eventMsg
			}
		}
	}
}

func (stream *Event) ServeHttp() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientChan := make(ClientChan)
		stream.NewClients <- clientChan

		defer func() {
			go func() {
				for range clientChan {
					// prevent block when send message to ch
				}
			}()
			stream.ClosedClients <- clientChan
		}()

		c.Set(ClientChanCtxKey, clientChan)
		c.Next()
	}
}

func SSeHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// set sse headers
		c.Writer.Header().Set("Content-Type", "text/event-stream")
		c.Writer.Header().Set("Cache-Control", "no-cache")
		c.Writer.Header().Set("Connection", "keep-alive")
		c.Writer.Header().Set("Transfer-Encoding", "chunked")

		c.Next()
	}
}

func GinSseServe() {
	router := gin.Default()

	stream := NewStreamServer()

	go func() {
		for {
			time.Sleep(2 * time.Second)
			now := time.Now().Format(time.DateTime)
			currentTime := fmt.Sprintf("The Current Time Is %v", now)
			stream.MessageChan <- currentTime
		}
	}()

	authorized := router.Group("/", gin.BasicAuth(gin.Accounts{
		"root": "jin",
	}))

	authorized.GET("/stream", SSeHeadersMiddleware(), stream.ServeHttp(), func(c *gin.Context) {
		v, ok := c.Get(ClientChanCtxKey)
		if !ok {
			return
		}

		clientChan, ok := v.(ClientChan)
		if !ok {
			return
		}

		// send stream message to client from message channel
		c.Stream(func(w io.Writer) bool {
			if msg, ok := <-clientChan; ok {
				c.SSEvent("message", msg)
				return true
			}
			return false
		})
	})

	router.StaticFile("/", "./static/index.html")

	router.Run(":8081")
}
