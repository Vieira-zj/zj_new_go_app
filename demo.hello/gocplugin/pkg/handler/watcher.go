package handler

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"

	"demo.hello/gocplugin/pkg"
	"demo.hello/utils"
	"github.com/gin-gonic/gin"
)

type watcherListSrvCoverReq struct {
	SrvName string `json:"srv_name" binding:"required"`
	Limit   int    `json:"limit" binding:"required"`
}

// ListSavedSrvCoversHandler .
func ListSavedSrvCoversHandler(c *gin.Context) {
	var param watcherListSrvCoverReq
	if err := c.ShouldBindJSON(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	if param.Limit < 1 {
		err := fmt.Errorf("Limit cannot be less than 1")
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	savedDirPath := getSavedCoverDirPath(param.SrvName)
	fileNames, err := utils.ListFilesInDir(savedDirPath, "cov")
	if err != nil {
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}
	sort.Strings(fileNames)

	limit := param.Limit
	if len(fileNames) < limit {
		limit = len(fileNames)
	}
	c.JSON(http.StatusOK, gin.H{"code": 0, "data": fileNames[len(fileNames)-limit:]})
}

type watcherGetSrvCoverReq struct {
	SrvName     string `json:"srv_name" binding:"required"`
	CovFileName string `json:"cov_file_name"`
}

// GetSrvCoverDataHandler .
func GetSrvCoverDataHandler(c *gin.Context) {
	var param watcherGetSrvCoverReq
	if err := c.ShouldBindJSON(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	var (
		covFilePath string
		err         error
	)

	savedDirPath := getSavedCoverDirPath(param.SrvName)
	if len(param.CovFileName) > 0 {
		covFilePath = filepath.Join(savedDirPath, param.CovFileName)
	} else {
		covFilePath, err = utils.GetLatestFileInDir(savedDirPath, "cov")
		if err != nil {
			sendErrorResp(c, http.StatusInternalServerError, err)
			return
		}
	}

	b, err := utils.ReadFile(covFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err = fmt.Errorf("Cov file not found: %s", param.CovFileName)
			sendErrorResp(c, http.StatusBadRequest, err)
		} else {
			sendErrorResp(c, http.StatusInternalServerError, err)
		}
		return
	}
	sendBytes(c, b)
}

// FetchAndSaveSrvCoverHandler 服务异常退出时调用该接口去拉取服务覆盖率数据，这里同步执行代替异步。
func FetchAndSaveSrvCoverHandler(c *gin.Context) {
	var param pkg.SyncSrvCoverParam
	if err := c.ShouldBindJSON(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	if len(param.Addresses) == 0 {
		err := fmt.Errorf("Addresses is empty")
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	savedDirPath := getSavedCoverDirPath(param.SrvName)
	if _, err := pkg.FetchAndSaveSrvCoverByAddr(savedDirPath, param); err != nil {
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}
	sendMessageResp(c, "Fetch and save service cover success.")
}

func getSavedCoverDirPath(srvName string) string {
	dir := pkg.GetSrvModuleDir(srvName)
	return filepath.Join(dir, pkg.WatcherCoverDataDirName)
}
