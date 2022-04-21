package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// IndexHandler .
func IndexHandler(c *gin.Context) {
	c.String(http.StatusOK, "Goc plugin: ok")
}

// PingHandler .
func PingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

//
// Helper
//

func sendBytes(c *gin.Context, body []byte) {
	c.Data(http.StatusOK, "application/octet-stream", body)
}

func sendMessageResp(c *gin.Context, msg string) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg})
}

func sendSrvCoverTotalResp(c *gin.Context, coverTotal string) {
	const msg = "Sync service cover success"
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": msg, "cover_total": coverTotal})
}

func sendErrorResp(c *gin.Context, errCode int, err error) {
	c.JSON(errCode, gin.H{"code": errCode, "error": err.Error()})
}
