package handler

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"demo.hello/gocplugin/pkg"
	"demo.hello/utils"
	"github.com/gin-gonic/gin"
)

type respSrvCoverItem struct {
	pkg.SyncSrvCoverParam
	CoverTotal string `json:"cover_total"`
}

// GetListOfSrvCoversHandler .
func GetListOfSrvCoversHandler(c *gin.Context) {
	gocAPI := pkg.NewGocAPI()
	srvs, err := gocAPI.ListRegisterServices(c.Request.Context())
	if err != nil {
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}

	srvList := make([]respSrvCoverItem, 0, len(srvs))
	for srvName, addrs := range srvs {
		item := respSrvCoverItem{}
		item.SrvName = srvName
		item.Addresses = addrs

		param := pkg.SyncSrvCoverParam{
			SrvName:   srvName,
			Addresses: addrs,
		}
		if item.CoverTotal, err = getLatestSrvCoverTotal(param); err != nil {
			sendErrorResp(c, http.StatusInternalServerError, err)
		}
		srvList = append(srvList, item)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "count": len(srvList), "data": srvList})
}

// GetLatestSrvCoverTotalHandler .
func GetLatestSrvCoverTotalHandler(c *gin.Context) {
	param := pkg.SyncSrvCoverParam{}
	if err := c.ShouldBindJSON(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	coverTotal, err := getLatestSrvCoverTotal(param)
	if err != nil {
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}
	sendSrvCoverTotalResp(c, coverTotal)
}

func getLatestSrvCoverTotal(param pkg.SyncSrvCoverParam) (string, error) {
	dbInstance := pkg.NewGocSrvCoverDBInstance()
	meta := pkg.GetSrvMetaFromName(param.SrvName)
	row, err := dbInstance.GetLatestSrvCoverRow(meta)
	if err != nil {
		if errors.Is(err, pkg.ErrSrvCoverLatestRowNotFound) {
			if len(param.Addresses) == 0 {
				if err := setSrvAddrsForParam(&param); err != nil {
					return "", fmt.Errorf("getLatestSrvCoverTotal error: %w", err)
				}
			}
			coverTotal, err := pkg.GetSrvTotalFromGoc(param.Addresses)
			if err != nil {
				return "", fmt.Errorf("getLatestSrvCoverTotal error: %w", err)
			}
			return coverTotal, nil
		}
		return "", fmt.Errorf("getLatestSrvCoverTotal error: %w", err)
	}

	return row.CoverTotal.String, nil
}

func setSrvAddrsForParam(param *pkg.SyncSrvCoverParam) error {
	if err := pkg.RemoveUnhealthSrvInGocTask(); err != nil {
		return fmt.Errorf("setSrvAddrsForParam error: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), pkg.ShortWait)
	defer cancel()

	gocAPI := pkg.NewGocAPI()
	srvs, err := gocAPI.ListRegisterServices(ctx)
	if err != nil {
		return fmt.Errorf("setSrvAddrsForParam error: %w", err)
	}

	addrs, ok := srvs[param.SrvName]
	if !ok {
		err = fmt.Errorf("Service name not found: %s", param.SrvName)
		return fmt.Errorf("setSrvAddrsForParam error: %w", err)
	}
	if len(addrs) == 0 {
		err = fmt.Errorf("No address found for service: %s", param.SrvName)
		return fmt.Errorf("setSrvAddrsForParam error: %w", err)
	}
	param.Addresses = addrs
	return nil
}

type respSrvCoverTotalItem struct {
	ID         uint
	Addresses  []string
	Commit     string
	CoverTotal string
}

// GetHistorySrvCoverTotalsHandler .
func GetHistorySrvCoverTotalsHandler(c *gin.Context) {
	param := pkg.SyncSrvCoverParam{}
	if err := c.ShouldBindJSON(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	dbInstance := pkg.NewGocSrvCoverDBInstance()
	meta := pkg.GetSrvMetaFromName(param.SrvName)
	rows, err := dbInstance.GetLimitedHistorySrvCoverRows(meta, 10)
	if err != nil {
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}

	srvCoverTotals := make([]respSrvCoverTotalItem, 0, len(rows))
	for _, row := range rows {
		item := respSrvCoverTotalItem{
			ID:         row.ID,
			Addresses:  strings.Split(row.Addrs, ","),
			Commit:     row.GitCommit,
			CoverTotal: row.CoverTotal.String,
		}
		srvCoverTotals = append(srvCoverTotals, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":             0,
		"srv_name":         meta.AppName,
		"count":            len(srvCoverTotals),
		"srv_cover_totals": srvCoverTotals,
	})
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
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}
	sendBytes(c, b)
}

// SyncSrvCoverHandler sync service cover data from goc, and create report.
func SyncSrvCoverHandler(c *gin.Context) {
	param := pkg.SyncSrvCoverParam{}
	if err := c.ShouldBindJSON(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}
	if err := setSrvAddrsForParam(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	tasksState := pkg.NewSrvCoverSyncTasksState()
	if srvState, ok := tasksState.Get(param.SrvName); ok {
		switch srvState {
		case pkg.StateRunning:
			sendMessageResp(c, "Sync service cover task is currently running.")
			return
		case pkg.StateFreshed:
			if getIsForceSync(c) {
				break
			}
			if coverTotal, err := getLatestSrvCoverTotal(param); err != nil {
				sendErrorResp(c, http.StatusInternalServerError, err)
			} else {
				sendSrvCoverTotalResp(c, coverTotal)
			}
			return
		}
	}

	retCh := pkg.SubmitSrvCoverSyncTask(param)
	select {
	case res := <-retCh:
		if coverTotal, ok := res.(string); ok {
			sendSrvCoverTotalResp(c, coverTotal)
		} else if err, ok := res.(error); ok {
			sendErrorResp(c, http.StatusInternalServerError, err)
		}
	case <-time.After(pkg.LongWait):
		sendErrorResp(c, http.StatusOK, fmt.Errorf("Timeout for wait sync service cover task"))
	}
}

func getIsForceSync(c *gin.Context) bool {
	force := c.Query("force")
	if strings.ToLower(force) != "true" {
		return false
	}
	return true
}

type getLatestCoverReportReq struct {
	RptType string `json:"rpt_type" binding:"required"`
	SrvName string `json:"srv_name" binding:"required"`
}

// GetLatestSrvCoverReportHandler .
func GetLatestSrvCoverReportHandler(c *gin.Context) {
	var req getLatestCoverReportReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	rptType := strings.ToLower(req.RptType)
	if rptType != "html" && rptType != "func" {
		err := fmt.Errorf("Invalid report type: %s", req.RptType)
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	meta := pkg.GetSrvMetaFromName(req.SrvName)
	dbInstance := pkg.NewGocSrvCoverDBInstance()
	row, err := dbInstance.GetLatestSrvCoverRow(meta)
	if err != nil {
		if errors.Is(err, pkg.ErrSrvCoverLatestRowNotFound) {
			sendErrorResp(c, http.StatusBadRequest, err)
		} else {
			sendErrorResp(c, http.StatusInternalServerError, err)
		}
		return
	}

	filePath := pkg.GetFilePathWithNewExt(row.CovFilePath, rptType)
	b, err := utils.ReadFile(filePath)
	if err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}
	sendBytes(c, b)
}
