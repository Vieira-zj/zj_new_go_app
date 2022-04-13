package handler

import (
	"fmt"
	"net/http"

	"demo.hello/gocadapter/pkg"
	"github.com/gin-gonic/gin"
)

type getSrvRawCoverRequest struct {
	SrvAddr string `json:"srv_addr" binding:"required"`
}

// GetSrvRawCoverHandler .
func GetSrvRawCoverHandler(c *gin.Context) {
	req := getSrvRawCoverRequest{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gocAPI := pkg.NewGocAPI(pkg.AppConfig.GocHost)
	b, err := gocAPI.GetServiceProfileByAddr(c.Request.Context(), req.SrvAddr)
	if err != nil {
		c.String(http.StatusInternalServerError, fmt.Sprintf("Get service profile for [%s] failed: %v", req.SrvAddr, err))
		return
	}
	c.Data(http.StatusOK, "application/octet-stream", b)
}
