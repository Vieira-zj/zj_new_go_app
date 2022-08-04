package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	"go1_1711_demo/middlewares/gin/pkg"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

/*
rest api:
curl -v http://127.0.0.1:8081/
curl -v http://127.0.0.1:8081/user
curl -v http://127.0.0.1:8081/users
curl -v http://127.0.0.1:8081/notfound

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

	r.NoMethod(HandleNotFound)
	r.NoRoute(HandleNotFound)

	r.Use(ErrHandler())

	r.GET("/", Ping)
	r.GET("/user", User)
	r.GET("/users", Users)

	// NOTE: if state value is not triggerred, it will not show in metrics results
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	pro := r.Group("/prometheus")

	// NOTE: must register middleware brefore router
	pro.Use(PrometheusHandler)

	pro.GET("/apia", PrometheusHandleA)
	pro.POST("/apib", PrometheusHandleB)

	if err := r.Run(":8081"); err != nil {
		log.Fatalln(err)
	}
}

// Handle

func HandleNotFound(c *gin.Context) {
	err := pkg.NotFound
	c.JSON(err.StatusCode, err)
	return
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

// Handle Mock Error

func User(c *gin.Context) {
	c.Error(fmt.Errorf("user name not found"))
}

func Users(c *gin.Context) {
	// in middleware, it try to rewrite status code and headers which set by AbortWithError, it gets warn:
	// [WARNING] Headers were already written. Wanted to override status code 200 with 500
	// here, use Error instead of AbortWithError.
	c.AbortWithError(http.StatusOK, pkg.ServerError)
}

// Prometheus Handle

func PrometheusHandleA(c *gin.Context) {
	sleep := rand.Intn(200)
	time.Sleep(time.Duration(sleep) * time.Millisecond)
	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "get ok",
	})
}

func PrometheusHandleB(c *gin.Context) {
	sleep := rand.Intn(300)
	time.Sleep(time.Duration(sleep) * time.Millisecond)
	c.AbortWithStatusJSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "post ok",
	})

}

// Middleware

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
