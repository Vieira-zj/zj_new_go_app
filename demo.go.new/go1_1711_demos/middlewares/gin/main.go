package main

import (
	"context"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

/*
rest api:
curl http://127.0.0.1:8081/
curl http://127.0.0.1:8081/ping

curl -v http://127.0.0.1:8081/notfound

curl -XPOST http://127.0.0.1:8081/data/copybody -d '{"id":101, "content":"body test"}'
curl http://127.0.0.1:8081/data/stream -v
curl http://127.0.0.1:8081/data/compress -H "Accept-Encoding: gzip" -v | gunzip

curl "http://127.0.0.1:8081/test/panic?trigger=true"
curl "http://127.0.0.1:8081/test/abort?type=none"

curl http://127.0.0.1:8081/test/ctxval

curl "http://127.0.0.1:8081/auth/login?name=foo"

curl -XPOST "http://127.0.0.1:8081/signup" -d '{"name":"foo","age":21,"date":"2023-09-03","email":"foo@gmail.com"}'

curl http://127.0.0.1:8081/error/user/bar
curl http://127.0.0.1:8081/error/users


api for metrics test:
curl http://127.0.0.1:8081/prometheus/apia
curl -XPOST http://127.0.0.1:8081/prometheus/apib

metrics api:
curl http://127.0.0.1:8081/metrics | grep api
*/

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()
	r.ContextWithFallback = true

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("checkDate", checkSignUpDate)
	}

	initRouter(r)

	srv := &http.Server{
		Addr:    ":8081",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			log.Fatal(err)
		}
	}()

	// kill -2 发送 syscall.SIGINT 信号，常用的 Ctrl+C 就是触发系统 SIGINT 信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("server exit")
}

func initRouter(r *gin.Engine) {
	r.NoMethod(notFoundHandler)
	r.NoRoute(notFoundHandler)

	r.GET("/", pingHandler)
	r.GET("/ping", pingHandler)

	data := r.Group("data")
	data.POST("/copybody", copyBodyHandler)
	data.GET("/stream", streamHandler)
	data.GET("/compress", compressHandler)

	test := r.Group("/test")
	test.GET("/panic", recoverMiddleware(), panicHandler)
	test.GET("/abort", logger1Middleware(), abortMiddleware(), logger2Middleware(), abortHandler)

	test.GET("/ctxval", contextMiddleware(), getContextValueHandler)

	r.POST("/signup", signUpHanler)
	auth := r.Group("/auth").Use(authMiddleware())
	auth.GET("/login", loginHandler)

	er := r.Group("error").Use(errMiddleware())
	er.GET("/user/:name", userHandler)
	er.GET("/users", usersHandler)

	// NOTE: if state value is not triggerred, it will not show in metrics results
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// NOTE: must register middleware brefore router
	pro := r.Group("/prometheus").Use(prometheusMiddleware)
	pro.GET("/apia", prometheusHandlerA)
	pro.POST("/apib", prometheusHandlerB)
}
