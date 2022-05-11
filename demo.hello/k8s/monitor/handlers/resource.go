package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"demo.hello/k8s/monitor/internal"
	"github.com/labstack/echo"
)

// PodFilter .
type PodFilter struct {
	NameSpaces []string `json:"namespaces"`
	Names      []string `json:"names"`
	IPs        []string `json:"ips"`
}

// ResponsePodInfo .
type ResponsePodInfo struct {
	Total int                   `json:"total"`
	Data  []*internal.PodStatus `json:"data"`
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

// GetPodsStatusByFilter return pods status by filter.
func GetPodsStatusByFilter(c echo.Context, lister *internal.Lister) error {
	body := c.Request().Body
	defer body.Close()
	b, err := ioutil.ReadAll(body)
	if err != nil {
		c.Logger().Error(err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	if len(b) == 0 {
		err := fmt.Errorf("request body is empty")
		c.Logger().Error(err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}

	filter := &PodFilter{}
	if err := json.Unmarshal(b, filter); err != nil {
		c.Logger().Error(err.Error())
		return c.String(http.StatusBadRequest, err.Error())
	}

	podsStatus, err := lister.GetAllPodInfosByListWatch(c.Request().Context())
	if err != nil {
		c.Logger().Error(err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	retStatus := filterPods(podsStatus, filter)

	return c.JSON(http.StatusOK, ResponsePodInfo{
		Total: len(retStatus),
		Data:  retStatus,
	})
}

func filterPods(pods []*internal.PodStatus, filter *PodFilter) []*internal.PodStatus {
	nsFilter := make(map[string]struct{}, len(filter.NameSpaces))
	for _, ns := range filter.NameSpaces {
		nsFilter[ns] = struct{}{}
	}
	ipFilter := make(map[string]struct{}, len(filter.IPs))
	for _, ip := range filter.IPs {
		ipFilter[ip] = struct{}{}
	}

	retPods := make([]*internal.PodStatus, 0)
	for _, status := range pods {
		if len(nsFilter) > 0 {
			if _, ok := nsFilter[status.Namespace]; !ok {
				continue
			}
		}
		if _, ok := ipFilter[status.IPAddress]; ok {
			retPods = append(retPods, status)
			continue
		}
		// 模糊匹配 name
		for _, name := range filter.Names {
			s := strings.ToLower(status.Name)
			substr := strings.ToLower(name)
			if strings.Contains(s, substr) {
				retPods = append(retPods, status)
				break
			}
		}
	}
	return retPods
}
