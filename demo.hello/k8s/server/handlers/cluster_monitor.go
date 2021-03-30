package handlers

import (
	"context"
	"net/http"

	k8spkg "demo.hello/k8s/client/pkg"
	"github.com/labstack/echo"
)

var resource *k8spkg.Resource

func init() {
	clientset, err := k8spkg.CreateK8sClient()
	if err != nil {
		panic(err)
	}
	resource = k8spkg.NewResource(context.TODO(), clientset)
}

// GetPodsStatus returns pods status: pod name, status, readiness and message.
func GetPodsStatus(c echo.Context) error {
	namespace := c.QueryParam("namespace")
	if len(namespace) == 0 {
		return c.String(http.StatusBadRequest, "namespace cannot be empty.")
	}

	resource.SetContext(c.Request().Context())
	podInfos, err := k8spkg.GetAllPodInfos(resource, namespace)
	if err != nil {
		c.Logger().Error(err.Error())
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, podInfos)
}
