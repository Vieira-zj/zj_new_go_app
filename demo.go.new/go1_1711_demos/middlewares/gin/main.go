package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
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

curl -XPOST http://127.0.0.1:8081/test/copybody -d '{"id":101, "content":"body test"}'
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
	r.NoMethod(NotFoundHandler)
	r.NoRoute(NotFoundHandler)

	r.GET("/", PingHandler)

	test := r.Group("/test")
	test.POST("/copybody", CopyBodyHandler)
	test.GET("/abort", logger1Middleware(), abortMiddleware(), logger2Middleware(), AbortHandler)
	test.GET("/ctxval", ContextMiddleware(), GetContextValueHandler)

	r.POST("/signup", SignUpHanler)
	auth := r.Group("/auth").Use(AuthMiddleware())
	auth.GET("/login", LoginHandler)

	er := r.Group("error").Use(ErrMiddleware())
	er.GET("/user/:name", UserHandler)
	er.GET("/users", UsersHandler)

	// NOTE: if state value is not triggerred, it will not show in metrics results
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// NOTE: must register middleware brefore router
	pro := r.Group("/prometheus").Use(PrometheusMiddleware)
	pro.GET("/apia", PrometheusHandlerA)
	pro.POST("/apib", PrometheusHandlerB)
}

//
// Handler
//

func NotFoundHandler(c *gin.Context) {
	err := pkg.NotFound
	c.JSON(err.StatusCode, err)
	return
}

func PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Pong",
	})
}

func LoginHandler(c *gin.Context) {
	log.Println("access login")
	name := c.Query("name")
	c.String(http.StatusOK, "welcome "+name)
}

//
// CopyBody Handler
//

type testBody struct {
	Id      int    `json:"id" binding:"required"`
	Content string `json:"content"`
}

func CopyBodyHandler(c *gin.Context) {
	body := c.Request.Body
	bodyBytes, err := ioutil.ReadAll(body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "read request body error: " + err.Error(),
		})
		return
	}
	log.Printf("post body: %s", bodyBytes)
	body.Close()

	// reset body
	c.Request.Body = ioutil.NopCloser(bytes.NewReader(bodyBytes))
	req := testBody{}
	if err = c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "parse json body error: " + err.Error(),
		})
		return
	}
	log.Printf("read json body: id=%d, content=[%s]", req.Id, req.Content)

	c.JSON(http.StatusOK, gin.H{
		"message": "cpoied ok",
	})
}

//
// Abort Handler
//
// normal:       logger1 -> abort -> logger2 -> handler -> logger2 -> abort -> logger1
// before abort: logger1 -> abort -> loggeer1
// after abort:  logger1 -> abort -> logger2 -> handler -> logger2 -> logger1
//

func AbortHandler(c *gin.Context) {
	log.Println("abort handler go")

	time.Sleep(100 * time.Millisecond)
	abortType := c.Request.URL.Query().Get("type")
	if abortType == "after" {
		c.Error(fmt.Errorf("mock error: abort after"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"msg": "abort handler finished",
	})
}

// Mock Error Handler

func UserHandler(c *gin.Context) {
	name := c.Param("name")
	c.Error(fmt.Errorf(fmt.Sprintf("user name [%s] not found", name)))
}

func UsersHandler(c *gin.Context) {
	c.Error(pkg.ServerError)
}

// Param Validate Handler

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

func SignUpHanler(c *gin.Context) {
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

// Prometheus Handler

func PrometheusHandlerA(c *gin.Context) {
	sleep := rand.Intn(200)
	time.Sleep(time.Duration(sleep) * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "apia get ok",
	})
}

func PrometheusHandlerB(c *gin.Context) {
	sleep := rand.Intn(300)
	time.Sleep(time.Duration(sleep) * time.Millisecond)
	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "apib post ok",
	})
}

// Wrap Context Handler

const ctxTestKey = "ctx-test-key"

func GetContextValueHandler(c *gin.Context) {
	// condition: enable r.ContextWithFallback
	val := c.Value(ctxTestKey)
	if val == nil {
		log.Println("no value in gin context")
	}
	log.Println("value in gin context:", val)

	ctx := c.Request.Context()
	val = ctx.Value(ctxTestKey)
	if val == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "no value in request context",
		})
		return
	}
	log.Println("value in request context:", val)

	c.JSON(http.StatusOK, gin.H{
		"code":      0,
		"ctx-value": val,
	})
}

//
// Middleware
//

func abortMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("abort middleware start")
		abortType := c.Request.URL.Query().Get("type")
		if abortType == "before" {
			time.Sleep(30 * time.Millisecond)
			c.AbortWithStatusJSON(http.StatusOK, gin.H{
				"message": "abort before",
			})
			return
		}

		c.Next()

		if len(c.Errors) == 1 {
			time.Sleep(50 * time.Millisecond)
			err := c.Errors[0]
			c.JSON(http.StatusOK, gin.H{
				"error": err.Error(),
			})
			return
		}
		log.Println("abort middleware end")
	}
}

func logger1Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("logger1 middleware start")
		start := time.Now()
		log.Printf("\tmethod:%s, path:%s, query:%s", c.Request.Method, c.Request.URL.Path, c.Request.URL.RawQuery)
		time.Sleep(50 * time.Millisecond)

		c.Next()

		duration := time.Since(start).Milliseconds()
		log.Printf("\tsince: %d millisecs", duration)
		log.Println("logger1 middleware done")
	}
}

func logger2Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("logger2 middleware start")
		start := time.Now()
		log.Printf("\tabort_type:%s", c.Request.URL.Query().Get("type"))

		c.Next()

		duration := time.Since(start).Milliseconds()
		log.Printf("\tsince: %d millisecs", duration)
		log.Println("logger2 middleware done")
	}
}

func ContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("wrap for reqeust context")
		ctx := c.Request.Context()
		newCtx := context.WithValue(ctx, ctxTestKey, "new-value")
		c.Request = c.Request.WithContext(newCtx)
		c.Next()
	}
}

// AuthMiddleware: auth verify handler.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		name := c.Query("name")
		if !strings.EqualFold(name, "foo") {
			err := pkg.AuthError(fmt.Sprintf("user name [%s] not found", name))
			log.Printf("status code: %d, message: %s", err.StatusCode, err.Msg)
			// middleware 中如果出现错误不想继续后续接口的调用不能直接使用 return, 而是应该调用 c.Abort() 方法
			// 因此这里使用 c.AbortWithStatusJSON() 代替 c.JSON()
			// c.JSON(err.StatusCode, err)
			c.AbortWithStatusJSON(err.StatusCode, err)
			return
		}
		log.Println("auth verify pass, and go next")
		c.Next()
	}
}

// ErrMiddleware: gin global error handler.
func ErrMiddleware() gin.HandlerFunc {
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

// PrometheusMiddleware: api monitor handler.
func PrometheusMiddleware(c *gin.Context) {
	start := time.Now()
	method := c.Request.Method
	pkg.GaugeVecApiMethod.WithLabelValues(method).Inc()

	c.Next()

	end := time.Now()
	d := end.Sub(start).Milliseconds()
	path := c.Request.URL.Path
	pkg.HistogramVecApiDuration.WithLabelValues(path).Observe(float64(d))
}
