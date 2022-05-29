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
	sendSuccessResp(c, "Pong")
}

//
// Common
//

func sendSuccessResp(c *gin.Context, msg string) {
	if len(msg) == 0 {
		msg = "Success."
	}
	sendResp(c, http.StatusOK, msg)
}

func sendErrorResp(c *gin.Context, errCode int, errMsg string) {
	sendResp(c, errCode, errMsg)
}

func sendResp(c *gin.Context, retCode int, message string) {
	c.JSON(retCode, gin.H{"code": retCode, "message": message})
}

func sendSyncSrvCoverResp(c *gin.Context, item respSrvCoverSubItem) {
	msg := "Sync service cover success"
	if item.SrvStatus == srvStatusOffline {
		msg = "Service is offline, return history cover value"
	}
	c.JSON(http.StatusOK, gin.H{
		"code":    http.StatusOK,
		"message": msg,
		"data":    item,
	})
}

func sendBytes(c *gin.Context, body []byte) {
	c.Data(http.StatusOK, "application/octet-stream", body)
}
