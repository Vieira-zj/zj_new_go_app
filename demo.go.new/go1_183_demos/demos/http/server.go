package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	s := FakeService{}
	mux := http.NewServeMux()
	mux.HandleFunc("/sleep", s.Handler)

	port := ":8081"
	srv := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	go func() {
		log.Println("serve listen at " + port)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http serve error: %v", err)
		}
		log.Println("stopped serving new connections")
	}()

	quit, stop := signal.NotifyContext(context.TODO(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()
	<-quit.Done()
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

type FakeService struct {
	wg sync.WaitGroup
}

// curl "http://localhost:8081/sleep?duration=3s"
func (s *FakeService) Handler(w http.ResponseWriter, r *http.Request) {
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
