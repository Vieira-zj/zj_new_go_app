package handler

import (
	"errors"
	"fmt"
	"net/http"
	"path/filepath"
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

	srvCoverItems := make([]respSrvCoverItem, 0, len(srvs))
	for srvName, addrs := range srvs {
		var item respSrvCoverItem
		if item.CoverTotal, err = getLatestSrvCoverTotal(srvName); err != nil {
			sendErrorResp(c, http.StatusInternalServerError, err)
			return
		}
		item.SrvName = srvName
		item.Addresses = addrs
		srvCoverItems = append(srvCoverItems, item)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "count": len(srvCoverItems), "data": srvCoverItems})
}

// GetLatestSrvCoverTotalHandler .
func GetLatestSrvCoverTotalHandler(c *gin.Context) {
	var param pkg.SyncSrvCoverParam
	if err := c.ShouldBindJSON(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}
	if err := pkg.IsSrvOKInGoc(param.SrvName); err != nil {
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}

	coverTotal, err := getLatestSrvCoverTotal(param.SrvName)
	if err != nil {
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "cover_total": coverTotal})
}

func getLatestSrvCoverTotal(srvName string) (string, error) {
	dbInstance := pkg.NewGocSrvCoverDBInstance()
	meta := pkg.GetSrvMetaFromName(srvName)
	row, err := dbInstance.GetLatestSrvCoverRow(meta)
	if err != nil {
		if errors.Is(err, pkg.ErrSrvCoverLatestRowNotFound) {
			return "0.00", nil
		}
		return "", fmt.Errorf("getLatestSrvCoverTotal error: %w", err)
	}

	return row.CoverTotal.String, nil
}

type respSrvCoverTotalItem struct {
	ID         uint
	Addresses  []string
	Commit     string
	CoverTotal string
}

// GetHistorySrvCoverTotalsHandler .
func GetHistorySrvCoverTotalsHandler(c *gin.Context) {
	var param pkg.SyncSrvCoverParam
	if err := c.ShouldBindJSON(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}
	if err := pkg.IsSrvOKInGoc(param.SrvName); err != nil {
		sendErrorResp(c, http.StatusInternalServerError, err)
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
	var param pkg.SyncSrvCoverParam
	if err := c.ShouldBindJSON(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}
	if err := pkg.IsSrvOKInGoc(param.SrvName); err != nil {
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}

	gocAPI := pkg.NewGocAPI()
	b, err := gocAPI.GetServiceProfileByName(c.Request.Context(), param.SrvName)
	if err != nil {
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}
	sendBytes(c, b)
}

type syncSrvCoverReq struct {
	pkg.SyncSrvCoverParam
	IsForce bool `json:"is_force"`
}

// SyncSrvCoverHandler sync service cover data from goc, and create report.
func SyncSrvCoverHandler(c *gin.Context) {
	var req syncSrvCoverReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	tasksState := pkg.NewSrvCoverSyncTasksState()
	if srvState, ok := tasksState.Get(req.SrvName); ok {
		switch srvState {
		case pkg.StateRunning:
			sendMessageResp(c, "Sync service cover task is currently running.")
			return
		case pkg.StateFreshed:
			if req.IsForce {
				break
			}
			if coverTotal, err := getLatestSrvCoverTotal(req.SrvName); err != nil {
				sendErrorResp(c, http.StatusInternalServerError, err)
			} else {
				sendSrvCoverTotalResp(c, coverTotal)
			}
			return
		}
	}

	retCh := pkg.SubmitSrvCoverSyncTask(pkg.SyncSrvCoverParam{
		SrvName:   req.SrvName,
		Addresses: req.Addresses,
	})
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

// GetLatestSrvFuncCoverReportHandler .
func GetLatestSrvFuncCoverReportHandler(c *gin.Context) {
	var param pkg.SyncSrvCoverParam
	if err := c.ShouldBindJSON(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	meta := pkg.GetSrvMetaFromName(param.SrvName)
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

	filePath := pkg.GetFilePathWithNewExt(row.CovFilePath, pkg.CoverRptTypeFunc)
	b, err := utils.ReadFile(filePath)
	if err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}
	sendBytes(c, b)
}

type listSrvCoverReportsReq struct {
	watcherListSrvCoverReq
	RptType string `json:"rpt_type" binding:"required"`
}

// ListSrvCoverReportsHandler .
func ListSrvCoverReportsHandler(c *gin.Context) {
	var req listSrvCoverReportsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	if req.RptType != pkg.CoverRptTypeFunc && req.RptType != pkg.CoverRptTypeHTML {
		err := fmt.Errorf("Invalid report type %s", req.RptType)
		sendErrorResp(c, http.StatusBadRequest, err)
	}

	if req.Limit < 1 {
		err := fmt.Errorf("Limit cannot be less than 1")
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	meta := pkg.GetSrvMetaFromName(req.SrvName)
	listDirPath := filepath.Join(pkg.AppConfig.RootDir, pkg.ReportCoverDataDirName)
	if req.RptType == pkg.CoverRptTypeHTML {
		listDirPath = filepath.Join(pkg.AppConfig.PublicDir, meta.AppName)
	}

	names, err := listFileNamesFromDir(listDirPath, pkg.CoverRptTypeHTML, req.Limit)
	if err != nil {
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": names})
}

// ClearSrvCoverHandler .
func ClearSrvCoverHandler(c *gin.Context) {
	// TODO:
}
