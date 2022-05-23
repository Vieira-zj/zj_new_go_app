package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"demo.hello/gocplugin/pkg"
	"demo.hello/utils"
	"github.com/gin-gonic/gin"
)

// ListAttachedSrvsHandler .
func ListAttachedSrvsHandler(c *gin.Context) {
	srvs, err := pkg.SyncAndListRegisterSrvsTask()
	if err != nil {
		log.Println("ListAttachedSrvsHandler error:", err)
		respErrMsg := "Sync and list services from goc register list failed."
		sendErrorResp(c, http.StatusInternalServerError, respErrMsg)
		return
	}

	avaliableAddrs := make([]string, 0, len(srvs))
	for _, addrs := range srvs {
		avaliableAddrs = append(avaliableAddrs, addrs...)
	}
	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "addresses": avaliableAddrs})
}

type watcherListSrvCoverReq struct {
	SrvName string `json:"srv_name" binding:"required"`
	Limit   int    `json:"limit" binding:"required"`
}

// ListSrvRawCoversHandler .
func ListSrvRawCoversHandler(c *gin.Context) {
	var param watcherListSrvCoverReq
	if err := c.ShouldBindJSON(&param); err != nil {
		log.Println("ListSrvRawCoversHandler error:", err)
		sendErrorResp(c, http.StatusBadRequest, errMsgJSONBind)
		return
	}

	if param.Limit < 1 {
		respErrMsg := "Limit cannot be less than 1."
		sendErrorResp(c, http.StatusBadRequest, respErrMsg)
		return
	}

	savedDirPath := getSavedCoverDirPath(param.SrvName)
	meta := pkg.GetSrvMetaFromName(param.SrvName)
	fileNames, err := listFileNamesFromDir(savedDirPath, "cov", meta.GitCommit, param.Limit)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			sendErrorResp(c, http.StatusBadRequest, "Service cov file not found.")
			return
		}
		log.Println("ListSrvRawCoversHandler error:", err)
		sendErrorResp(c, http.StatusInternalServerError, "List file failed.")
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "cover_files:": fileNames})
}

func listFileNamesFromDir(dirPath, fileExt, filter string, limit int) ([]string, error) {
	if !utils.IsDirExist(dirPath) {
		return nil, os.ErrNotExist
	}

	fileNames, err := utils.ListFilesInDir(dirPath, fileExt)
	if err != nil {
		return nil, fmt.Errorf("listFileNamesFromDir error: %w", err)
	}

	filterFileNames := fileNames
	if len(filter) > 0 {
		filterFileNames = make([]string, 0, len(fileNames))
		for _, name := range fileNames {
			if strings.Contains(name, filter) {
				filterFileNames = append(filterFileNames, name)
			}
		}
	}
	sort.Strings(filterFileNames)

	if len(filterFileNames) < limit {
		limit = len(filterFileNames)
	}
	return filterFileNames[len(filterFileNames)-limit:], nil
}

type watcherGetSrvRawCoverReq struct {
	SrvName     string `json:"srv_name" binding:"required"`
	CovFileName string `json:"cov_file_name"`
}

// GetSrvRawCoverHandler .
func GetSrvRawCoverHandler(c *gin.Context) {
	var req watcherGetSrvRawCoverReq
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Println("GetSrvRawCoverHandler error:", err)
		sendErrorResp(c, http.StatusBadRequest, errMsgJSONBind)
		return
	}

	var (
		covFileName string
		err         error
	)

	savedDirPath := getSavedCoverDirPath(req.SrvName)
	if len(req.CovFileName) > 0 {
		covFileName = req.CovFileName
	} else {
		covFileName, err = utils.GetLatestFileInDir(savedDirPath, "cov")
		if err != nil {
			log.Println("GetSrvRawCoverHandler error:", err)
			respErrMsg := "Get latest file in dir failed."
			sendErrorResp(c, http.StatusInternalServerError, respErrMsg)
			return
		}
	}

	b, err := utils.ReadFile(filepath.Join(savedDirPath, covFileName))
	if err != nil {
		log.Println("GetSrvRawCoverHandler error:", err)
		if errors.Is(err, os.ErrNotExist) {
			sendErrorResp(c, http.StatusBadRequest, "Cov file is not exist.")
		} else {
			sendErrorResp(c, http.StatusInternalServerError, "Read file failed.")
		}
		return
	}
	sendBytes(c, b)
}

// SyncSrvCoverHookHandler fetch and save service raw cover.
// 服务异常退出时调用该接口去拉取服务覆盖率数据，这里同步执行代替异步。
func SyncSrvCoverHookHandler(c *gin.Context) {
	var param pkg.SyncSrvCoverParam
	if err := c.ShouldBindJSON(&param); err != nil {
		log.Println("SyncSrvCoverHookHandler error:", err)
		sendErrorResp(c, http.StatusBadRequest, errMsgJSONBind)
		return
	}

	ok, err := isSrvExistInGocList(param.SrvName)
	if err != nil {
		log.Println("SyncSrvCoverHookHandler error:", err)
		sendErrorResp(c, http.StatusInternalServerError, errMsgCheckSrvInGocList)
		return
	}
	if !ok {
		sendErrorResp(c, http.StatusBadRequest, errMsgSrvNotExist)
		return
	}

	savedDirPath := getSavedCoverDirPath(param.SrvName)
	if !utils.IsDirExist(savedDirPath) {
		if err := utils.MakeDir(savedDirPath); err != nil {
			log.Println("SyncSrvCoverHookHandler error:", err)
			sendErrorResp(c, http.StatusInternalServerError, "Make dir failed.")
			return
		}
	}

	if _, err := pkg.FetchAndSaveSrvCover(savedDirPath, param.SrvName); err != nil {
		log.Println("SyncSrvCoverHookHandler error:", err)
		respErrMsg := "Sync service cover failed."
		sendErrorResp(c, http.StatusInternalServerError, respErrMsg)
		return
	}
	sendSuccessResp(c, "Sync service raw cover success.")
}

func getSavedCoverDirPath(srvName string) string {
	dir := pkg.GetSrvModuleDir(srvName)
	return filepath.Join(dir, pkg.WatcherCoverDataDirName)
}
