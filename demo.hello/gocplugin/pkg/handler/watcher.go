package handler

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"

	"demo.hello/gocplugin/pkg"
	"demo.hello/utils"
	"github.com/gin-gonic/gin"
)

type watcherListSrvCoverReq struct {
	SrvName string
	Limit   uint
}

// ListSavedSrvCoversHandler .
func ListSavedSrvCoversHandler(c *gin.Context) {
	param := watcherListSrvCoverReq{}
	if err := c.ShouldBindJSON(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	savedDirPath := getSavedCoverDirPath(param.SrvName)
	fileNames, err := utils.ListFilesInDir(savedDirPath, "cov")
	if err != nil {
		sendErrorResp(c, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"code": 0, "data": fileNames[param.Limit]})
}

type watcherGetSrvCoverReq struct {
	SrvName     string
	CovFileName string
}

// GetSrvCoverDataHandler .
func GetSrvCoverDataHandler(c *gin.Context) {
	param := watcherGetSrvCoverReq{}
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
		sendErrorResp(c, http.StatusInternalServerError, err)
	}
	sendBytes(c, b)
}

// FetchAndSaveSrvCoverHandler .
func FetchAndSaveSrvCoverHandler(c *gin.Context) {
	param := pkg.SyncSrvCoverParam{}
	if err := c.ShouldBindJSON(&param); err != nil {
		sendErrorResp(c, http.StatusBadRequest, err)
		return
	}

	go func() {
		savedDirPath := getSavedCoverDirPath(param.SrvName)
		if _, err := pkg.FetchAndSaveSrvCover(savedDirPath, param); err != nil {
			ctx, cancel := context.WithTimeout(context.Background(), pkg.Wait)
			defer cancel()
			notify := pkg.NewMatterMostNotify()
			errMsg := fmt.Sprintln("FetchAndSaveSrvCoverHandler error:", err)
			notify.MustSendMessageToDefaultUser(ctx, errMsg)
		}
	}()

	sendMessageResp(c, "Trigger fetch and save srv cover success.")
}

func getSavedCoverDirPath(srvName string) string {
	dir := pkg.GetModuleDir(srvName)
	return filepath.Join(dir, pkg.WatcherCoverDataDirName)
}
