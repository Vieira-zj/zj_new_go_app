package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// HTML rendering
	r := gin.Default()

	// Multiple templates
	r.LoadHTMLGlob("templates/**/*")
	// curl http://localhost:8081/posts/index
	r.GET("/posts/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "posts/index.tmpl", gin.H{
			"title": "Posts",
		})
	})

	// curl http://localhost:8081/users/index
	r.GET("/users/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "users/index.tmpl", gin.H{
			"title": "Users",
		})
	})

	// Redirects
	// Redirect to external location
	// chrome: http://localhost:8081/test
	r.GET("/test", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "http://www.google.com/")
	})

	// Redirect to internal location
	// curl http://localhost:8081/test2 | jq .
	r.GET("/test2", func(c *gin.Context) {
		c.Request.URL.Path = "/test3"
		r.HandleContext(c)
	})

	r.GET("/test3", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hello"})
	})

	r.Run(":8081")
}
