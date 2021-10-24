package handlers

import (
	"net/http"

	"demo.hello/k8s/monitor/internal"
	"github.com/labstack/echo"
)

// ResponsePodInfo .
type ResponsePodInfo struct {
	Total int `json:"total"`
	Data  []*internal.PodStatus
}

// GetPodsStatus returns pods status: pod name, status, message and log.
func GetPodsStatus(c echo.Context, lister *internal.Lister) error {
	podInfos, err := lister.GetAllPodInfosByRaw(c.Request().Context())
	if err != nil {
		c.Logger().Error(err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ResponsePodInfo{
		Total: len(podInfos),
		Data:  podInfos,
	})
}

// GetPodsStatusByList returns pods status by list watcher.
func GetPodsStatusByList(c echo.Context, lister *internal.Lister) error {
	podInfos, err := lister.GetAllPodInfosByListWatch(c.Request().Context())
	if err != nil {
		c.Logger().Error(err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, ResponsePodInfo{
		Total: len(podInfos),
		Data:  podInfos,
	})
}
