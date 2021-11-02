package main

import (
	"log"
	"net/http"
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

func main() {
	// Custom Middleware
	r := gin.New()
	r.Use(MyLogger())

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
			log.Println("Done in path:", ctxCopied.Request.URL.Path)
		}()
		c.String(http.StatusOK, "ok")
	})

	// curl http://localhost:8081/long_sync
	r.GET("long_sync", func(c *gin.Context) {
		time.Sleep(3 * time.Second)
		log.Println("Done in path:", c.Request.URL.Path)
		c.String(http.StatusOK, "ok")
	})

	r.Run(":8081")
}
