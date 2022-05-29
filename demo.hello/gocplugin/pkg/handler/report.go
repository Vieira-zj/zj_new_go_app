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
	errMsgGetSrvAddrs         = "Get service addresses failed."

	srvStatusOnline  = "online"
	srvStatusOffline = "offline"
)

//
// Cover Total
//

type respSrvCoverSubItem struct {
	SrvStatus  string `json:"srv_status"`
	CoverTotal string `json:"cover_total"`
	UpdatedAt  string `json:"updated_at"`
}

type respSrvCoverItem struct {
	pkg.SyncSrvCoverParam
	respSrvCoverSubItem
}

// ListOfSrvCoversHandler .
func ListOfSrvCoversHandler(c *gin.Context) {
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
		if item.CoverTotal, item.UpdatedAt, err = getLatestSrvCoverTotalInDB(srvName); err != nil {
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

	coverTotal, updatedAt, err := getLatestSrvCoverTotalInDB(param.SrvName)
	if err != nil {
		log.Println("GetLatestSrvCoverTotalHandler error:", err)
		sendErrorResp(c, http.StatusInternalServerError, errMsgGetLatestSrvCovInDB)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": http.StatusOK,
		"data": respSrvCoverSubItem{
			SrvStatus:  srvStatus,
			CoverTotal: coverTotal,
			UpdatedAt:  updatedAt,
		},
	})
}

const emptyDate = "null"

func getLatestSrvCoverTotalInDB(srvName string) (string, string, error) {
	dbInstance := pkg.NewGocSrvCoverDBInstance()
	meta := pkg.GetSrvMetaFromName(srvName)
	row, err := dbInstance.GetLatestSrvCoverRow(meta)
	if err != nil {
		if errors.Is(err, pkg.ErrSrvCoverLatestRowNotFound) {
			return pkg.ZeroCoverTotal, emptyDate, nil
		}
		return "", emptyDate, fmt.Errorf("getLatestSrvCoverTotal error: %w", err)
	}

	return row.CoverTotal.String, pkg.GetSimpleDatetime(row.UpdatedAt), nil
}

type respHistorySrvCoverItem struct {
	Addresses  []string `json:"addresses"`
	Commit     string   `json:"commit"`
	CoverTotal string   `json:"cover_total"`
	UpdatedAt  string   `json:"updated_at"`
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

	const defaultLimit = 10
	dbInstance := pkg.NewGocSrvCoverDBInstance()
	meta := pkg.GetSrvMetaFromName(param.SrvName)
	rows, err := dbInstance.GetLimitedHistorySrvCoverRows(meta, defaultLimit)
	if err != nil {
		log.Println("GetHistorySrvCoverTotalsHandler error:", err)
		respErrMsg := "Get history service cover in db failed."
		sendErrorResp(c, http.StatusInternalServerError, respErrMsg)
		return
	}

	srvCoverTotals := make([]respHistorySrvCoverItem, 0, len(rows))
	for _, row := range rows {
		item := respHistorySrvCoverItem{
			Addresses:  strings.Split(row.Addrs, ","),
			Commit:     row.GitCommit,
			CoverTotal: row.CoverTotal.String,
			UpdatedAt:  pkg.GetSimpleDatetime(row.UpdatedAt),
		}
		srvCoverTotals = append(srvCoverTotals, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"code":       http.StatusOK,
		"srv_name":   meta.AppName,
		"srv_status": srvStatus,
		"count":      len(srvCoverTotals),
		"data":       srvCoverTotals,
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

	addrs, err := getSrvIPAddresses(req.SrvName)
	if err != nil {
		log.Println("SyncSrvCoverHandler error:", err)
		sendErrorResp(c, http.StatusInternalServerError, errMsgGetSrvAddrs)
		return
	}

	if len(addrs) == 0 {
		item := respSrvCoverSubItem{
			SrvStatus: srvStatusOffline,
		}
		item.CoverTotal, item.UpdatedAt, err = getLatestSrvCoverTotalInDB(req.SrvName)
		if err != nil {
			log.Println("SyncSrvCoverHandler error:", err)
			sendErrorResp(c, http.StatusInternalServerError, errMsgGetLatestSrvCovInDB)
			return
		}
		sendSyncSrvCoverResp(c, item)
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
			item := respSrvCoverSubItem{
				SrvStatus: srvStatusOnline,
			}
			if item.CoverTotal, item.UpdatedAt, err = getLatestSrvCoverTotalInDB(req.SrvName); err != nil {
				log.Println("SyncSrvCoverHandler error:", err)
				sendErrorResp(c, http.StatusInternalServerError, errMsgGetLatestSrvCovInDB)
			} else {
				sendSyncSrvCoverResp(c, item)
			}
			return
		}
	}

	retCh := pkg.SubmitSrvCoverSyncTask(pkg.SyncSrvCoverParam{
		SrvName:   req.SrvName,
		Addresses: addrs,
	})
	select {
	case res := <-retCh:
		if coverTotal, ok := res.(string); ok {
			sendSyncSrvCoverResp(c, respSrvCoverSubItem{
				SrvStatus:  srvStatusOnline,
				CoverTotal: coverTotal,
				UpdatedAt:  pkg.GetSimpleDatetime(time.Now()),
			})
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

	names, err := listFileNamesFromDir(listDirPath, req.RptType, meta.GitCommit, req.Limit)
	if err != nil {
		log.Println("ListSrvCoverReportsHandler error:", err)
		sendErrorResp(c, http.StatusInternalServerError, "List file failed.")
		return
	}
	if len(names) == 0 {
		sendErrorResp(c, http.StatusBadRequest, "No reports found.")
		return
	}

	retNames := make([]string, 0, len(names))
	for _, name := range names {
		retNames = append(retNames, filepath.Join(meta.AppName, name))
	}
	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "reports": retNames})
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
	addrs, err := getSrvIPAddresses(srvName)
	if err != nil {
		return false, err
	}

	if len(addrs) > 0 {
		return true, nil
	}
	return false, nil
}

func getSrvIPAddresses(srvName string) ([]string, error) {
	srvs, err := pkg.SyncAndListRegisterSrvsTask()
	if err != nil {
		return nil, fmt.Errorf("IsSrvOKInGoc error: %w", err)
	}

	for srv, addrs := range srvs {
		if srv == srvName {
			return addrs, nil
		}
	}
	return []string{}, nil
}

// IsSrvExistInWatcher .
func IsSrvExistInWatcher(srvName string) error {
	// TODO:
	return nil
}
