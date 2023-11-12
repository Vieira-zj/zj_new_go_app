package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// init function can be used for setup when the plugin is loaded.
func init() {
	fmt.Println("Greeter plugin loaded!")

	go runServe()
}

// Greeter is an exported variable, which will be accessible in the plugin.
var Greeter string = "Hello, World!"

// Greet is an exported function, which will be callable in the plugin.
func Greet(name string) string {
	return fmt.Sprintf("%s, %s!", Greeter, name)
}

func runServe() {
	gin.SetMode(gin.DebugMode)
	r := gin.Default()

	r.GET("/healthz", func(ctx *gin.Context) {
		ctx.String(http.StatusOK, "ok")
	})

	log.Println("Start server at :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
