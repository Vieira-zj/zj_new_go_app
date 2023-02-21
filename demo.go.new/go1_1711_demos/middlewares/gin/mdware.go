package main

import (
	"context"
	"fmt"
	"go1_1711_demo/middlewares/gin/pkg"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func recoverMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				c.AbortWithStatusJSON(http.StatusOK, gin.H{
					"error": fmt.Sprintf("recover from error: %v", err),
				})
			}
		}()

		log.Println("recover middleware")
		c.Next()
	}
}

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

func contextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Println("wrap for reqeust context")
		ctx := c.Request.Context()
		newCtx := context.WithValue(ctx, ctxTestKey, "new-value")
		c.Request = c.Request.WithContext(newCtx)
		c.Next()
	}
}

// authMiddleware: auth verify handler.
func authMiddleware() gin.HandlerFunc {
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

// errMiddleware: gin global error handler.
func errMiddleware() gin.HandlerFunc {
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

// prometheusMiddleware: api monitor handler.
func prometheusMiddleware(c *gin.Context) {
	start := time.Now()
	method := c.Request.Method
	pkg.GaugeVecApiMethod.WithLabelValues(method).Inc()

	c.Next()

	end := time.Now()
	d := end.Sub(start).Milliseconds()
	path := c.Request.URL.Path
	pkg.HistogramVecApiDuration.WithLabelValues(path).Observe(float64(d))
}
