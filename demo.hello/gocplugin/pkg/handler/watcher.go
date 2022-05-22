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

// ListCoverAttachSrvsHandler .
func ListCoverAttachSrvsHandler(c *gin.Context) {
	// TODO:
}

type watcherListSrvCoverReq struct {
	SrvName string `json:"srv_name" binding:"required"`
	Limit   int    `json:"limit" binding:"required"`
}

// ListSrvRawCoversHandler .
func ListSrvRawCoversHandler(c *gin.Context) {
	var param watcherListSrvCoverReq
	if err := c.ShouldBindJSON(&param); err != nil {
		log.Println("ListSavedSrvCoversHandler error:", err)
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
		log.Println("ListSavedSrvCoversHandler error:", err)
		sendErrorResp(c, http.StatusInternalServerError, "List file failed.")
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": fileNames})
}

func listFileNamesFromDir(dirPath, fileExt, filter string, limit int) ([]string, error) {
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

type watcherGetSrvCoverReq struct {
	SrvName     string `json:"srv_name" binding:"required"`
	CovFileName string `json:"cov_file_name"`
}

// GetSrvRawCoverHandler .
func GetSrvRawCoverHandler(c *gin.Context) {
	var param watcherGetSrvCoverReq
	if err := c.ShouldBindJSON(&param); err != nil {
		log.Println("GetSrvCoverHandler error:", err)
		sendErrorResp(c, http.StatusBadRequest, errMsgJSONBind)
		return
	}

	var (
		covFileName string
		err         error
	)

	savedDirPath := getSavedCoverDirPath(param.SrvName)
	if len(param.CovFileName) > 0 {
		covFileName = param.CovFileName
	} else {
		covFileName, err = utils.GetLatestFileInDir(savedDirPath, "cov")
		if err != nil {
			log.Println("GetSrvCoverHandler error:", err)
			respErrMsg := "Get latest file in dir failed."
			sendErrorResp(c, http.StatusInternalServerError, respErrMsg)
			return
		}
	}

	b, err := utils.ReadFile(filepath.Join(savedDirPath, covFileName))
	if err != nil {
		log.Println("GetSrvCoverHandler error:", err)
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
		log.Println("FetchAndSaveSrvCoverHandler error:", err)
		sendErrorResp(c, http.StatusBadRequest, errMsgJSONBind)
		return
	}

	if len(param.Addresses) == 0 {
		sendErrorResp(c, http.StatusBadRequest, "IP addresses is empty")
		return
	}

	savedDirPath := getSavedCoverDirPath(param.SrvName)
	if _, err := pkg.FetchAndSaveSrvCoverByAddr(savedDirPath, param); err != nil {
		log.Println("FetchAndSaveSrvCoverHandler error:", err)
		respErrMsg := "Fetch and save service cover failed."
		sendErrorResp(c, http.StatusInternalServerError, respErrMsg)
		return
	}
	sendSuccessResp(c, "Fetch and save service cover success")
}

func getSavedCoverDirPath(srvName string) string {
	dir := pkg.GetSrvModuleDir(srvName)
	return filepath.Join(dir, pkg.WatcherCoverDataDirName)
}
