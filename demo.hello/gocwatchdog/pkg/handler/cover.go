package handler

import (
	"fmt"
	"net/http"

	"demo.hello/gocwatchdog/pkg"
	"github.com/gin-gonic/gin"
)

type getSrvRawCoverReq struct {
	SrvAddr string `json:"srv_addr" binding:"required"`
}

// GetSrvRawCoverHandler .
func GetSrvRawCoverHandler(c *gin.Context) {
	req := getSrvRawCoverReq{}
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

type getLatestCoverReportReq struct {
	RptType string `json:"rpt_type" binding:"required"`
	Env     string `json:"env" binding:"required"`
	Region  string `json:"region" binding:"required"`
	AppName string `json:"app_name" binding:"required"`
}

// GetLatestCoverReportHandler .
func GetLatestCoverReportHandler(c *gin.Context) {
	var req getLatestCoverReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	dbInstance := pkg.NewGocSrvCoverDBInstance()
	row, err := dbInstance.GetLatestSrvCoverRow(pkg.SrvCoverMeta{
		Env:     req.Env,
		Region:  req.Region,
		AppName: req.AppName,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	funcRptPath := pkg.GetFilePathWithNewExt(row.CovFilePath, "func")
	coverTotal, err := pkg.GetCoverTotalFromFuncReport(funcRptPath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"cover_total": coverTotal})
}
