package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

/*
rest api:

curl -v http://127.0.0.1:8081/
curl -v http://127.0.0.1:8081/user
curl -v http://127.0.0.1:8081/users
curl -v http://127.0.0.1:8081/notfound
*/

func main() {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	r.NoMethod(HandleNotFound)
	r.NoRoute(HandleNotFound)
	r.Use(ErrHandler())

	r.GET("/", Ping)
	r.GET("/user", User)
	r.GET("/users", Users)

	if err := r.Run(":8081"); err != nil {
		log.Fatalln(err)
	}
}

// Handle

func HandleNotFound(c *gin.Context) {
	err := NotFound
	c.JSON(err.StatusCode, err)
	return
}

func Ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}

func User(c *gin.Context) {
	c.Error(fmt.Errorf("user name not found"))
}

func Users(c *gin.Context) {
	// in middleware, it try to rewrite status code and headers which set by AbortWithError, it gets warn:
	// [WARNING] Headers were already written. Wanted to override status code 200 with 500
	// here, use Error instead of AbortWithError.
	c.AbortWithError(http.StatusOK, ServerError)
}

// Middleware

// ErrHandler: gin 全局错误处理
func ErrHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
		if length := len(c.Errors); length > 0 {
			e := c.Errors[length-1]
			if err := e.Err; err != nil {
				var Err *Error
				if e, ok := err.(*Error); ok {
					Err = e
				} else if e, ok := err.(error); ok {
					Err = AuthError(e.Error())
				} else {
					Err = ServerError
				}
				c.Header("Content-Type", "application/json")
				c.JSON(Err.StatusCode, Err)
				return
			}
		}
	}
}
