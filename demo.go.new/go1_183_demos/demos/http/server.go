package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	s := FakeService{}
	mux := http.NewServeMux()
	mux.HandleFunc("/sleep", s.SleepHandler)
	mux.HandleFunc("/cancel", CancelHandler)

	port := ":8081"
	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	srv.RegisterOnShutdown(func() {
		log.Println("register shutdown hook")
	})

	go func() {
		log.Println("serve listen at " + port)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http serve error: %v", err)
		}
		log.Println("stopped serving new connections")
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-quit
	log.Println("shutdown server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("http server shutdown error: %v", err)
	}

	s.GracefulStop(ctx)
	log.Println("http server graceful shutdown completed")
}

// Handler

// curl "http://localhost:8081/cancel"
func CancelHandler(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		log.Printf("duration: %.2f", time.Since(start).Seconds())
	}()

	ch := make(chan bool)
	go func() {
		time.Sleep(3 * time.Second)
		ch <- true
		close(ch)
	}()

	ctx := r.Context()

	select {
	case <-ch:
		fmt.Fprintln(w, "hello")
	case <-ctx.Done():
		log.Println("request cancelled")
	}
}

// Sleep Handler

type FakeService struct {
	wg sync.WaitGroup
}

// curl "http://localhost:8081/sleep?duration=3s"
func (s *FakeService) SleepHandler(w http.ResponseWriter, r *http.Request) {
	duration, err := time.ParseDuration(r.FormValue("duration"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	time.Sleep(duration)
	// 模拟需要异步执行的代码，比如注册接口异步发送邮件、发送 Kafka 消息等
	s.FakeSendEmail()

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "welcome http server")
}

func (s *FakeService) FakeSendEmail() {
	s.wg.Add(1)

	go func() {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("recovered panic: %v\n", err)
			}
			s.wg.Done()
		}()

		log.Println("mail goroutine enter")
		time.Sleep(5 * time.Second)
		log.Println("mail goroutine exit")
	}()
}

func (s *FakeService) GracefulStop(ctx context.Context) {
	log.Println("waiting for service to finish")
	quit := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(quit)
	}()

	select {
	case <-ctx.Done():
		log.Println("context was marked as done earlier, than user service has stopped")
	case <-quit:
		log.Println("service finished")
	}
}
