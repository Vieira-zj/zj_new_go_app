package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// XML, JSON, YAML and ProtoBuf rendering
	r := gin.Default()

	// JSON rendering
	// curl http://localhost:8081/someJSON | jq .
	r.GET("/someJSON", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
	})

	// curl http://localhost:8081/moreJSON | jq .
	r.GET("/moreJSON", func(c *gin.Context) {
		var msg struct {
			Name    string `json:"user"`
			Message string
			Number  int
		}

		msg.Name = "Foo"
		msg.Message = "hey"
		msg.Number = 123
		c.JSON(http.StatusOK, msg)
	})

	// XML rendering
	// curl http://localhost:8081/someXML
	r.GET("/someXML", func(c *gin.Context) {
		c.XML(http.StatusOK, gin.H{"message": "hey", "status": http.StatusOK})
	})

	// Serving static files
	// chrome: http://localhost:8081/static
	// curl http://localhost:8081/gin.log
	r.StaticFS("/static", http.Dir("/tmp/test"))
	r.StaticFile("/gin.log", "/tmp/test/gin.log")

	// Serving data from reader
	// chrome: http://localhost:8081/someDataFromReader
	r.GET("someDataFromReader", func(c *gin.Context) {
		response, err := http.Get("https://raw.githubusercontent.com/gin-gonic/logo/master/color.png")
		if err != nil || response.StatusCode != http.StatusOK {
			c.Status(http.StatusServiceUnavailable)
			return
		}

		reader := response.Body
		defer reader.Close()
		contentLength := response.ContentLength
		contentType := response.Header.Get("Content-Type")
		extraHeaders := map[string]string{
			"Content-Disposition": `attachment; filename="gopher.png"`,
		}
		c.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
	})

	// HTML rendering
	// curl http://localhost:8081/index
	r.LoadHTMLFiles("templates/index.tmpl")
	r.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Main website",
		})
	})

	r.Run(":8081")
}
