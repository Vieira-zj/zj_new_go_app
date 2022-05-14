package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"demo.hello/gocplugin/pkg"
	"demo.hello/utils"
	"github.com/gin-gonic/gin"
)

const (
	errMsgJSONBind            = "Json parameter bind error."
	errMsgGetLatestSrvCovInDB = "Get latest service cover in db failed."
	errMsgSrvNotExist         = "Service is not exist in goc register list."
	errMsgCheckSrvInGocList   = "Check service in goc list failed."
	errMsgGetSrvStatus        = "Get service status failed."
)

//
// Cover Total
//

type respSrvCoverItem struct {
	pkg.SyncSrvCoverParam
	SrvStatus  string `json:"srv_status"`
	CoverTotal string `json:"cover_total"`
}

// GetListOfSrvCoversHandler .
func GetListOfSrvCoversHandler(c *gin.Context) {
	srvs, err := pkg.SyncAndListRegisterSrvsTask()
	if err != nil {
		log.Println("GetListOfSrvCoversHandler error:", err)
		respErrMsg := "Sync and list services from goc register list failed."
		sendErrorResp(c, http.StatusInternalServerError, respErrMsg)
		return
	}

	srvCoverItems := make([]respSrvCoverItem, 0, len(srvs))
	for srvName, addrs := range srvs {
		var item respSrvCoverItem
		if item.CoverTotal, err = getLatestSrvCoverTotalInDB(srvName); err != nil {
			log.Println("GetListOfSrvCoversHandler error:", err)
			sendErrorResp(c, http.StatusInternalServerError, errMsgGetLatestSrvCovInDB)
			return
		}
		item.SrvName = srvName
		item.Addresses = addrs
		item.SrvStatus = srvStatusOnline
		srvCoverItems = append(srvCoverItems, item)
	}
	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "count": len(srvCoverItems), "data": srvCoverItems})
}

const (
	srvStatusOnline  = "online"
	srvStatusOffline = "offline"
)

// GetLatestSrvCoverTotalHandler .
func GetLatestSrvCoverTotalHandler(c *gin.Context) {
	var param pkg.SyncSrvCoverParam
	if err := c.ShouldBindJSON(&param); err != nil {
		log.Println("GetLatestSrvCoverTotalHandler error:", err)
		sendErrorResp(c, http.StatusBadRequest, errMsgJSONBind)
		return
	}

	srvStatus, err := getSrvStatus(param.SrvName)
	if err != nil {
		log.Println("GetLatestSrvCoverTotalHandler error:", err)
		sendErrorResp(c, http.StatusInternalServerError, errMsgGetSrvStatus)
		return
	}

	coverTotal, err := getLatestSrvCoverTotalInDB(param.SrvName)
	if err != nil {
		log.Println("GetLatestSrvCoverTotalHandler error:", err)
		sendErrorResp(c, http.StatusInternalServerError, errMsgGetLatestSrvCovInDB)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":        http.StatusOK,
		"srv_status":  srvStatus,
		"cover_total": coverTotal,
	})
}

func getLatestSrvCoverTotalInDB(srvName string) (string, error) {
	dbInstance := pkg.NewGocSrvCoverDBInstance()
	meta := pkg.GetSrvMetaFromName(srvName)
	row, err := dbInstance.GetLatestSrvCoverRow(meta)
	if err != nil {
		if errors.Is(err, pkg.ErrSrvCoverLatestRowNotFound) {
			return pkg.ZeroCoverTotal, nil
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
		log.Println("GetHistorySrvCoverTotalsHandler error:", err)
		sendErrorResp(c, http.StatusBadRequest, errMsgJSONBind)
		return
	}

	srvStatus, err := getSrvStatus(param.SrvName)
	if err != nil {
		log.Println("GetHistorySrvCoverTotalsHandler error:", err)
		sendErrorResp(c, http.StatusInternalServerError, errMsgGetSrvStatus)
		return
	}

	dbInstance := pkg.NewGocSrvCoverDBInstance()
	meta := pkg.GetSrvMetaFromName(param.SrvName)
	rows, err := dbInstance.GetLimitedHistorySrvCoverRows(meta, 10)
	if err != nil {
		log.Println("GetHistorySrvCoverTotalsHandler error:", err)
		respErrMsg := "Get history service cover in db failed."
		sendErrorResp(c, http.StatusInternalServerError, respErrMsg)
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
		"code":             http.StatusOK,
		"srv_name":         meta.AppName,
		"srv_status":       srvStatus,
		"count":            len(srvCoverTotals),
		"srv_cover_totals": srvCoverTotals,
	})
}

//
// Cover
//

// FetchSrvRawCoverHandler .
func FetchSrvRawCoverHandler(c *gin.Context) {
	var param pkg.SyncSrvCoverParam
	if err := c.ShouldBindJSON(&param); err != nil {
		log.Println("FetchSrvRawCoverHandler error:", err)
		sendErrorResp(c, http.StatusBadRequest, errMsgJSONBind)
		return
	}

	ok, err := isSrvExistInGocList(param.SrvName)
	if err != nil {
		log.Println("FetchSrvRawCoverHandler error:", err)
		sendErrorResp(c, http.StatusInternalServerError, errMsgCheckSrvInGocList)
		return
	}
	if !ok {
		sendErrorResp(c, http.StatusBadRequest, errMsgSrvNotExist)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), pkg.Wait)
	defer cancel()

	gocAPI := pkg.NewGocAPI()
	b, err := gocAPI.GetServiceProfileByName(ctx, param.SrvName)
	if err != nil {
		log.Println("FetchSrvRawCoverHandler error:", err)
		respErrMsg := "Get service profile failed."
		sendErrorResp(c, http.StatusInternalServerError, respErrMsg)
		return
	}
	sendBytes(c, b)
}

type syncSrvCoverReq struct {
	pkg.SyncSrvCoverParam
	IsForce bool `json:"is_force"`
}

// ClearSrvCoverHandler .
func ClearSrvCoverHandler(c *gin.Context) {
	var param pkg.SyncSrvCoverParam
	if err := c.ShouldBindJSON(&param); err != nil {
		log.Println("ClearSrvCoverHandler error:", err)
		sendErrorResp(c, http.StatusBadRequest, errMsgJSONBind)
		return
	}

	ok, err := isSrvExistInGocList(param.SrvName)
	if err != nil {
		log.Println("ClearSrvCoverHandler error:", err)
		sendErrorResp(c, http.StatusInternalServerError, errMsgCheckSrvInGocList)
		return
	}
	if !ok {
		sendErrorResp(c, http.StatusBadRequest, errMsgSrvNotExist)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), pkg.Wait)
	defer cancel()

	gocAPI := pkg.NewGocAPI()
	if _, err := gocAPI.ClearProfileServiceByName(ctx, param.SrvName); err != nil {
		log.Println("ClearSrvCoverHandler error:", err)
		respErrMsg := "Clear service cover failed."
		sendErrorResp(c, http.StatusInternalServerError, respErrMsg)
		return
	}

	tasksState := pkg.NewSrvCoverSyncTasksState()
	tasksState.Delete(param.SrvName)

	db := pkg.NewGocSrvCoverDBInstance()
	meta := pkg.GetSrvMetaFromName(param.SrvName)
	meta.Addrs = strings.Join(param.Addresses, ",")
	row := pkg.GocSrvCoverModel{
		SrvCoverMeta: meta,
		IsLatest:     true,
		CovFilePath:  pkg.CovFilePathNullValue,
		CoverTotal: sql.NullString{
			String: pkg.ZeroCoverTotal,
			Valid:  true,
		},
	}
	if err := db.AddLatestSrvCoverRow(row); err != nil {
		log.Println("ClearSrvCoverHandler error:", err)
		respErrMsg := "Add latest service cover row failed."
		sendErrorResp(c, http.StatusInternalServerError, respErrMsg)
		return
	}

	sendSuccessResp(c, "Cover clear success.")
}

// SyncSrvCoverHandler sync service cover data from goc, and create report.
func SyncSrvCoverHandler(c *gin.Context) {
	var req syncSrvCoverReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("SyncSrvCoverHandler error:", err)
		sendErrorResp(c, http.StatusBadRequest, errMsgJSONBind)
		return
	}

	srvStatus, err := getSrvStatus(req.SrvName)
	if err != nil {
		log.Println("SyncSrvCoverHandler error:", err)
		sendErrorResp(c, http.StatusInternalServerError, errMsgGetSrvStatus)
		return
	}

	if srvStatus == srvStatusOffline {
		coverTotal, err := getLatestSrvCoverTotalInDB(req.SrvName)
		if err != nil {
			log.Println("SyncSrvCoverHandler error:", err)
			sendErrorResp(c, http.StatusInternalServerError, errMsgGetLatestSrvCovInDB)
			return
		}
		sendSrvCoverTotalResp(c, srvStatus, coverTotal)
		return
	}

	tasksState := pkg.NewSrvCoverSyncTasksState()
	if srvState, ok := tasksState.Get(req.SrvName); ok {
		switch srvState {
		case pkg.StateRunning:
			sendSuccessResp(c, "Sync service cover task is currently running")
			return
		case pkg.StateFreshed:
			if req.IsForce {
				break
			}
			if coverTotal, err := getLatestSrvCoverTotalInDB(req.SrvName); err != nil {
				log.Println("SyncSrvCoverHandler error:", err)
				sendErrorResp(c, http.StatusInternalServerError, errMsgGetLatestSrvCovInDB)
			} else {
				sendSrvCoverTotalResp(c, srvStatusOnline, coverTotal)
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
			sendSrvCoverTotalResp(c, srvStatusOnline, coverTotal)
		} else if err, ok := res.(error); ok {
			log.Println("SyncSrvCoverHandler error:", err)
			respErrMsg := "Sync service cover task failed."
			sendErrorResp(c, http.StatusInternalServerError, respErrMsg)
		}
		close(retCh)
	case <-time.After(pkg.LongWait):
		respMsg := "Sync service cover task submitted and running."
		sendSuccessResp(c, respMsg)
	}
}

//
// Report
//

type listSrvCoverReportsReq struct {
	watcherListSrvCoverReq
	RptType string `json:"rpt_type" binding:"required"`
}

// ListSrvCoverReportsHandler .
func ListSrvCoverReportsHandler(c *gin.Context) {
	var req listSrvCoverReportsReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("ListSrvCoverReportsHandler error:", err)
		sendErrorResp(c, http.StatusBadRequest, errMsgJSONBind)
		return
	}

	if req.RptType != pkg.CoverRptTypeFunc && req.RptType != pkg.CoverRptTypeHTML {
		respErrMsg := fmt.Sprintf("Invalid parameter: rpt_type=%s", req.RptType)
		sendErrorResp(c, http.StatusBadRequest, respErrMsg)
		return
	}

	if req.Limit < 1 {
		respErrMsg := "Limit cannot be less than 1."
		sendErrorResp(c, http.StatusBadRequest, respErrMsg)
		return
	}

	meta := pkg.GetSrvMetaFromName(req.SrvName)
	listDirPath := filepath.Join(pkg.AppConfig.RootDir, meta.AppName, pkg.ReportCoverDataDirName)
	if req.RptType == pkg.CoverRptTypeHTML {
		listDirPath = filepath.Join(pkg.AppConfig.PublicDir, meta.AppName)
	}

	names, err := listFileNamesFromDir(listDirPath, req.RptType, req.Limit)
	if err != nil {
		log.Println("ListSrvCoverReportsHandler error:", err)
		sendErrorResp(c, http.StatusInternalServerError, "List file failed.")
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "reports": names})
}

type getSrvFuncCoverRptReq struct {
	pkg.SyncSrvCoverParam
	RptName string `json:"rpt_name"`
	RptType string `json:"rpt_type"`
}

// GetSrvCoverReportHandler returns cov or func cover report. (html report is get from static.)
func GetSrvCoverReportHandler(c *gin.Context) {
	var req getSrvFuncCoverRptReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("GetSrvCoverReportHandler error:", err)
		sendErrorResp(c, http.StatusBadRequest, errMsgJSONBind)
		return
	}

	if req.RptType != pkg.CoverRptTypeFunc && req.RptType != pkg.CoverRptTypeRaw {
		respErrMsg := fmt.Sprintf("Invalid parameter: rpt_type=%s", req.RptType)
		sendErrorResp(c, http.StatusBadRequest, respErrMsg)
		return
	}

	var (
		filePath string
		err      error
	)
	meta := pkg.GetSrvMetaFromName(req.SrvName)
	if len(req.RptName) > 0 {
		suffix := "." + req.RptType
		if !strings.HasSuffix(req.RptName, suffix) {
			req.RptName = req.RptName + suffix
		}
		filePath = filepath.Join(pkg.GetModuleCoverDataDir(meta.AppName), req.RptName)
	} else {
		// if rpt_name not set, get from latest row in db
		filePath, err = getLatestSrvCoverRptPath(meta, req.RptType)
		if err != nil {
			log.Println("GetSrvCoverReportHandler error:", err)
			if errors.Is(err, pkg.ErrSrvCoverLatestRowNotFound) {
				sendErrorResp(c, http.StatusBadRequest, "Not exist in db.")
			} else {
				sendErrorResp(c, http.StatusInternalServerError, "Get report path in db failed.")
			}
			return
		}
	}

	b, err := utils.ReadFile(filePath)
	if err != nil {
		log.Println("GetSrvCoverReportHandler error:", err)
		sendErrorResp(c, http.StatusBadRequest, "Read file failed.")
		return
	}
	sendBytes(c, b)
}

func getLatestSrvCoverRptPath(meta pkg.SrvCoverMeta, ext string) (string, error) {
	dbInstance := pkg.NewGocSrvCoverDBInstance()
	row, err := dbInstance.GetLatestSrvCoverRow(meta)
	if err != nil {
		return "", fmt.Errorf("getLatestSrvCoverRow error: %w", err)
	}
	retFilePath := pkg.FormatFilePathWithNewExt(row.CovFilePath, ext)
	return retFilePath, nil
}

//
// Common
//

func getSrvStatus(srvName string) (string, error) {
	ok, err := isSrvExistInGocList(srvName)
	if err != nil {
		return srvStatusOffline, fmt.Errorf("getSrvStatus error: %w", err)
	}

	if ok {
		return srvStatusOnline, nil
	}
	return srvStatusOffline, nil
}

func isSrvExistInGocList(srvName string) (bool, error) {
	srvs, err := pkg.SyncAndListRegisterSrvsTask()
	if err != nil {
		return false, fmt.Errorf("IsSrvOKInGoc error: %w", err)
	}

	for srv := range srvs {
		if srv == srvName {
			return true, nil
		}
	}
	return false, nil
}

// IsSrvExistInWatcher .
func IsSrvExistInWatcher(srvName string) error {
	// TODO:
	return nil
}
