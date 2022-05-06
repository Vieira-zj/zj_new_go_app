package main

import (
	"log"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
)

// MyLogger .
func MyLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// before request
		t := time.Now()
		c.Set("example", "12345")

		c.Next()

		// after request
		latency := time.Since(t)
		log.Println("latency:", latency)
		status := c.Writer.Status()
		log.Println("status:", status)
	}
}

// MyRecover .
func MyRecover(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("panic: %v", r)
			debug.PrintStack()
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": "Internal Server Error",
			})
		}
	}()
	c.Next()
}

func main() {
	// Custom Middlewares
	r := gin.New()
	r.Use(MyLogger())
	r.Use(MyRecover)

	// curl http://localhost:8081/test
	r.GET("/test", func(c *gin.Context) {
		example := c.MustGet("example").(string)
		log.Println("context [example] value:", example)
		c.String(http.StatusOK, "ok")
	})

	// Goroutines inside a middleware
	// curl http://localhost:8081/long_async
	r.GET("/long_async", func(c *gin.Context) {
		ctxCopied := c.Copy()
		go func() {
			time.Sleep(3 * time.Second)
			log.Println("done in path:", ctxCopied.Request.URL.Path)
		}()
		c.String(http.StatusOK, "ok")
	})

	// curl http://localhost:8081/long_sync
	r.GET("long_sync", func(c *gin.Context) {
		time.Sleep(3 * time.Second)
		log.Println("done in path:", c.Request.URL.Path)
		c.String(http.StatusOK, "ok")
	})

	r.Run(":8081")
}
