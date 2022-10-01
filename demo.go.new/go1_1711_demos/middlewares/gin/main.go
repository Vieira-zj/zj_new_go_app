package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go1_1711_demo/middlewares/gin/pkg"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

/*
rest api:
curl -v http://127.0.0.1:8081/
curl -v http://127.0.0.1:8081/notfound

curl -v "http://127.0.0.1:8081/auth/login?name=foo"

curl -XPOST "http://127.0.0.1:8081/signup" -d '{"name":"foo","age":21,"date":"2023-09-03","email":"foo@gmail.com"}'

curl -v http://127.0.0.1:8081/error/user/bar
curl -v http://127.0.0.1:8081/error/users

api for metrics test:
curl -v http://127.0.0.1:8081/prometheus/apia
curl -v -XPOST http://127.0.0.1:8081/prometheus/apib

metrics api:
curl http://127.0.0.1:8081/metrics | grep api
*/

func init() {
	rand.Seed(time.Now().Unix())
}

func main() {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

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
	r.NoMethod(HandleNotFound)
	r.NoRoute(HandleNotFound)

	r.GET("/", Ping)
	r.POST("/signup", SignUp)

	auth := r.Group("/auth").Use(AuthHandler())
	auth.GET("/login", Login)

	er := r.Group("error").Use(ErrHandler())
	er.GET("/user/:name", User)
	er.GET("/users", Users)

	// NOTE: if state value is not triggerred, it will not show in metrics results
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// NOTE: must register middleware brefore router
	pro := r.Group("/prometheus").Use(PrometheusHandler)
	pro.GET("/apia", PrometheusHandleA)
	pro.POST("/apib", PrometheusHandleB)
}

//
// Handle
//

func HandleNotFound(c *gin.Context) {
	err := pkg.NotFound
	c.JSON(err.StatusCode, err)
	return
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Pong",
	})
}

func Login(c *gin.Context) {
	log.Println("access login")
	name := c.Query("name")
	c.String(http.StatusOK, "welcome "+name)
}

// Handle Mock Error

func User(c *gin.Context) {
	name := c.Param("name")
	c.Error(fmt.Errorf(fmt.Sprintf("user name [%s] not found", name)))
}

func Users(c *gin.Context) {
	c.Error(pkg.ServerError)
}

// Handle Param Validate

type SignUpParam struct {
	Age   uint8  `json:"age" binding:"gte=1,lte=130"`
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Date  string `json:"date" binding:"required,datetime=2006-01-02,checkDate"`
}

func checkSignUpDate(fl validator.FieldLevel) bool {
	date, err := time.Parse("2006-01-02", fl.Field().String())
	if err != nil {
		return false
	}
	if date.Before(time.Now()) {
		return false
	}
	return true
}

func SignUp(c *gin.Context) {
	var s SignUpParam
	if err := c.ShouldBindJSON(&s); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	log.Printf("sign up: name=%s, age=%d, email=%s, date=%s", s.Name, s.Age, s.Email, s.Date)
	c.JSON(http.StatusOK, gin.H{
		"message": "signup success",
	})
}

// Handle Prometheus

func PrometheusHandleA(c *gin.Context) {
	sleep := rand.Intn(200)
	time.Sleep(time.Duration(sleep) * time.Millisecond)
	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "apia get ok",
	})
}

func PrometheusHandleB(c *gin.Context) {
	sleep := rand.Intn(300)
	time.Sleep(time.Duration(sleep) * time.Millisecond)
	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "apib post ok",
	})
}

//
// Middleware
//

// AuthHandler: auth verify handler.
func AuthHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("name")
		if !strings.EqualFold(name, "foo") {
			err := pkg.AuthError(fmt.Sprintf("user name [%s] not found", name))
			log.Printf("status code: %d, message: %s", err.StatusCode, err.Msg)
			// middleware 中如果出现错误不想继续后续接口的调用不能直接使用 return, 而是应该调用 c.Abort() 方法
			// 因此这里要使用 c.AbortWithStatusJSON() 代替 c.JSON()
			// c.JSON(err.StatusCode, err)
			c.AbortWithStatusJSON(err.StatusCode, err)
			return
		}
		log.Println("auth verify pass, and go next")
		c.Next()
	}
}

// ErrHandler: gin global error handler.
func ErrHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if length := len(c.Errors); length > 0 {
			e := c.Errors[length-1]
			if err := e.Err; err != nil {
				path := c.Request.URL.Path
				pkg.GaugeVecApiError.WithLabelValues(path).Inc()

				var Err *pkg.Error
				if e, ok := err.(*pkg.Error); ok {
					Err = e
				} else if e, ok := err.(error); ok {
					Err = pkg.AuthError(e.Error())
				} else {
					Err = pkg.ServerError
				}
				// c.Next() 已经执行完成，这里并没有使用 c.AbortWithStatusJSON(), 而是直接使用 c.JSON() 后 return
				c.Header("Content-Type", "application/json")
				c.JSON(Err.StatusCode, Err)
				return
			}
		}
	}
}

// PrometheusHandler: api monitor handler.
func PrometheusHandler(c *gin.Context) {
	start := time.Now()
	method := c.Request.Method
	pkg.GaugeVecApiMethod.WithLabelValues(method).Inc()

	c.Next()
	end := time.Now()
	d := end.Sub(start).Milliseconds()
	path := c.Request.URL.Path
	pkg.HistogramVecApiDuration.WithLabelValues(path).Observe(float64(d))
}
