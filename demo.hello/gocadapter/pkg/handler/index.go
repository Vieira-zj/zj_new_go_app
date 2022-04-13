package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// IndexHandler .
func IndexHandler(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

// PingHandler .
func PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}
