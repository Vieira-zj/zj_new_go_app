package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.New()

	// Write log file
	gin.DisableConsoleColor()
	f, _ := os.Create("/tmp/test/gin.log")
	gin.DefaultWriter = io.MultiWriter(os.Stdout, f)

	// Using middleware
	r.Use(gin.Logger())

	// Custom Log Format
	r.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	}))

	// Custom Recovery behavior
	r.Use(gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.String(http.StatusInternalServerError, fmt.Sprintf("error: %s", err))
		}
		c.AbortWithStatus(http.StatusInternalServerError)
	}))

	// curl http://localhost:8081/panic
	r.GET("/panic", func(c *gin.Context) {
		panic("foo")
	})

	// curl http://localhost:8081/
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "hello")
	})

	// curl http://localhost:8081/ping
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	r.Run(":8081")
}
