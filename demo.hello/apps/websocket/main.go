package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"demo.hello/apps/websocket/handlers"
)

func main() {
	server := http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: time.Duration(3) * time.Second,
		ReadTimeout:       time.Duration(5) * time.Second,
		WriteTimeout:      time.Duration(8) * time.Second,
	}

	http.HandleFunc("/", handlers.IndexHandler)
	http.HandleFunc("/mock/jobs", handlers.GetAllJobResultsHandler)
	http.HandleFunc("/mock/jobs/init", handlers.InitJobResultsHandler)
	http.HandleFunc("/ws/echo", handlers.EchoMessage)
	http.HandleFunc("/ws/jobs/sync", handlers.SyncDeltaJobResults)

	fmt.Println("http serve at :8080")
	go server.ListenAndServe()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	<-ctx.Done()

	stop()
	fmt.Println("shutting down gracefully, press Ctrl+C again to force")
	handlers.EventBus.Stop()
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		panic(err)
	}
}
