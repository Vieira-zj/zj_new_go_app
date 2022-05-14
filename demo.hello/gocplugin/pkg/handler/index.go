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

func sendSrvCoverTotalResp(c *gin.Context, srvStatus, coverTotal string) {
	msg := "Sync service cover success"
	if srvStatus == srvStatusOffline {
		msg = "Service is offline, return history cover value"
	}
	c.JSON(http.StatusOK, gin.H{
		"code":        http.StatusOK,
		"message":     msg,
		"srv_status":  srvStatus,
		"cover_total": coverTotal,
	})
}

func sendBytes(c *gin.Context, body []byte) {
	c.Data(http.StatusOK, "application/octet-stream", body)
}
