package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// curl http://localhost:8081/ping | jq .
	r.GET("ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Parameters in path
	// curl http://localhost:8081/user/foo
	r.GET("/user/:name", func(c *gin.Context) {
		name := c.Param("name")
		c.String(http.StatusOK, "Hello %s", name)
	})

	// curl http://localhost:8081/user/bar/send
	r.GET("/user/:name/*action", func(c *gin.Context) {
		log.Println("req path:", c.FullPath())
		name := c.Param("name")
		action := c.Param("action")
		message := name + " is " + action
		c.String(http.StatusOK, message)
	})

	// Query string parameters
	// curl "http://localhost:8081/welcome?firstname=Foo&lastname=Bar"
	r.GET("/welcome", func(c *gin.Context) {
		firstname := c.DefaultQuery("firstname", "Guest")
		lastname := c.Query("lastname")
		c.String(http.StatusOK, "Hello %s %s", firstname, lastname)
	})

	// Upload single file
	// curl -X POST http://localhost:8081/upload -H "Content-Type: multipart/form-data" \
	//   -F "file=@/Users/jinzheng/Downloads/tmps/convert.py"
	r.MaxMultipartMemory = 8 << 20 // 8 MiB
	r.POST("/upload", func(c *gin.Context) {
		file, _ := c.FormFile("file")
		log.Println("upload file:", file.Filename)
		c.SaveUploadedFile(file, "/tmp/test/"+file.Filename)
		c.String(http.StatusOK, fmt.Sprintf("'%s' uploaded", file.Filename))
	})

	// Grouping routes
	// curl http://localhost:8081/v1/version
	v1 := r.Group("/v1")
	{
		v1.GET("version", func(c *gin.Context) {
			c.String(http.StatusOK, "api version v1")
		})
	}

	// curl http://localhost:8081/v2/version
	v2 := r.Group("/v2")
	{
		v2.GET("version", func(c *gin.Context) {
			c.String(http.StatusOK, "api version v2")
		})
	}

	r.Run(":8081")
}
