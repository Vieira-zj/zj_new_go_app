package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response .
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// C .
var C *gin.Context

// Error .
func Error(message string) {
	if len(message) == 0 {
		message = "fail"
	}
	C.JSON(http.StatusOK, Response{
		Code:    99,
		Message: message,
		Data:    nil,
	})
}

// Success .
func Success(data interface{}) {
	C.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}
