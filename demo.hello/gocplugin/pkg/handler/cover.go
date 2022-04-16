package handler

import (
	"errors"
	"fmt"
	"net/http"

	"demo.hello/gocplugin/pkg"
	"demo.hello/utils"
	"github.com/gin-gonic/gin"
)

// SrvListItem .
type SrvListItem struct {
	pkg.SyncSrvCoverParam
	CoverTotal string `json:"cover_total"`
}

// GetCoverSrvListHandler .
func GetCoverSrvListHandler(c *gin.Context) {
	gocAPI := pkg.NewGocAPI()
	srvs, err := gocAPI.ListRegisterServices(c.Request.Context())
	if err != nil {
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}

	srvList := make([]SrvListItem, 0, len(srvs))
	dbInstance := pkg.NewGocSrvCoverDBInstance()
	for srvName, addrs := range srvs {
		item := SrvListItem{}
		item.SrvName = srvName
		item.Addresses = addrs

		meta := pkg.GetSrvMetaFromName(srvName)
		if row, err := dbInstance.GetLatestSrvCoverRow(meta); err != nil {
			if errors.Is(err, pkg.ErrSrvCoverLatestRowNotFound) {
				// get cover total from frist addr by default
				defaultAddr := addrs[0]
				if item.CoverTotal, err = pkg.GetSrvTotalFromGoc(defaultAddr); err != nil {
					sendErrorResp(c, http.StatusInternalServerError, err)
					return
				}
			} else {
				sendErrorResp(c, http.StatusInternalServerError, err)
				return
			}
		} else {
			item.CoverTotal = row.CoverTotal.String
		}
		srvList = append(srvList, item)
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "count": len(srvList), "data": srvList})
}

// GetSrvRawCoverHandler .
func GetSrvRawCoverHandler(c *gin.Context) {
	req := struct {
		SrvAddr string `json:"srv_addr" binding:"required"`
	}{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gocAPI := pkg.NewGocAPI()
	b, err := gocAPI.GetServiceProfileByAddr(c.Request.Context(), req.SrvAddr)
	if err != nil {
		err = fmt.Errorf("Get service profile error: %v", err)
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}
	sendBytes(c, b)
}

// SyncCoverReportHandler .
func SyncCoverReportHandler(c *gin.Context) {
	param := pkg.SyncSrvCoverParam{}
	if err := c.ShouldBindJSON(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	// TODO: use pool to run task
	if err := pkg.GetSrvCoverAndCreateReportTask(param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}
	sendMessageResp(c, "sync success")
}

type getLatestCoverReportReq struct {
	RptType string `json:"rpt_type" binding:"required"`
	SrvName string `json:"srv_name" binding:"required"`
}

// GetLatestCoverReportHandler .
func GetLatestCoverReportHandler(c *gin.Context) {
	var req getLatestCoverReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	meta := pkg.GetSrvMetaFromName(req.SrvName)
	dbInstance := pkg.NewGocSrvCoverDBInstance()
	row, err := dbInstance.GetLatestSrvCoverRow(meta)
	if err != nil {
		if errors.Is(err, pkg.ErrSrvCoverLatestRowNotFound) {
			err = fmt.Errorf("Cover report is not found for [%v]", req.SrvName)
			sendErrorResp(c, http.StatusBadRequest, err)
		} else {
			sendErrorResp(c, http.StatusInternalServerError, err)
		}
		return
	}

	filePath := pkg.GetFilePathWithNewExt(row.CovFilePath, req.RptType)
	b, err := utils.ReadFile(filePath)
	if err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}
	sendBytes(c, b)
}
