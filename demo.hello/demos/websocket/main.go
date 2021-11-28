package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"demo.hello/demos/websocket/handlers"
)

/*
api test:
curl http://localhost:8080/
curl http://localhost:8080/ping
*/

func main() {
	router := http.NewServeMux()
	router.HandleFunc("/", handlers.Index)
	router.HandleFunc("/ping", handlers.Ping)
	router.HandleFunc("/ws/echo", handlers.Echo)

	eventBusSvr := handlers.CreateEventBusServer()
	defer eventBusSvr.Stop()

	server := http.Server{
		Addr:              ":8080",
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      8 * time.Second,
		Handler:           trace(logging(router)),
	}

	fmt.Println("[main]: http serve at :8080")
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			panic(fmt.Errorf("server error: %v", err))
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	go handlers.MockMessage(ctx)

	<-ctx.Done()
	stop()

	fmt.Println("[main]: server shutdown")
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		fmt.Println("[main]: server shutdown error:", err)
	}
	time.Sleep(time.Second)
}

/*
aop
*/

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("logging")
		next.ServeHTTP(w, r)
	})
}

func trace(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("trace")
		next.ServeHTTP(w, r)
	})
}
