package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"demo.hello/utils"
)

const (
	// CoverProfileAPI .
	CoverProfileAPI = "/v1/cover/profile"
	// CoverProfileClearAPI .
	CoverProfileClearAPI = "/v1/cover/clear"
	// CoverServicesListAPI .
	CoverServicesListAPI = "/v1/cover/list"
	// CoverRegisterServiceAPI .
	CoverRegisterServiceAPI = "/v1/cover/register"
	// CoverServicesRemoveAPI .
	CoverServicesRemoveAPI = "/v1/cover/remove"
)

// GocParam .
type GocParam struct {
	Service []string `json:"service,omitempty"`
	Address []string `json:"address,omitempty"`
}

// GocAPI adapter for goc api.
type GocAPI struct {
	host string
	http *utils.HTTPUtils
}

var (
	gocAPI     *GocAPI
	gocAPIOnce sync.Once
)

// NewGocAPI .
func NewGocAPI(gocCenterHost string) *GocAPI {
	gocAPIOnce.Do(func() {
		gocAPI = &GocAPI{
			host: gocCenterHost,
			http: utils.NewDefaultHTTPUtils(),
		}
	})
	return gocAPI
}

//
// Service API
//

// ListRegisterServices .
func (goc *GocAPI) ListRegisterServices(ctx context.Context) (map[string][]string, error) {
	url := goc.host + CoverServicesListAPI
	resp, err := goc.http.Get(ctx, url, map[string]string{})
	if err != nil {
		return nil, fmt.Errorf("ListRegisterServices send http get failed: %w", err)
	}

	services := make(map[string][]string, 8)
	if err := json.Unmarshal(resp, &services); err != nil {
		return nil, fmt.Errorf("ListRegisterServices json unmarshal failed: %w", err)
	}
	return services, nil
}

// RegisterService .
func (goc *GocAPI) RegisterService(ctx context.Context, service, addr string) (string, error) {
	url := goc.host + CoverRegisterServiceAPI
	query := fmt.Sprintf("name=%s&address=%s", service, addr)
	resp, err := goc.http.Post(ctx, url+"?"+query, map[string]string{}, "")
	if err != nil {
		return "", fmt.Errorf("RegisterService send http post error: %w", err)
	}
	return string(resp), nil
}

// DeleteRegisterServiceByName .
func (goc *GocAPI) DeleteRegisterServiceByName(ctx context.Context, service string) (string, error) {
	return goc.deleteRegisterService(ctx, []string{service}, nil)
}

// DeleteRegisterServiceByAddr .
func (goc *GocAPI) DeleteRegisterServiceByAddr(ctx context.Context, addr string) (string, error) {
	return goc.deleteRegisterService(ctx, nil, []string{addr})
}

func (goc *GocAPI) deleteRegisterService(ctx context.Context, service, addr []string) (string, error) {
	param := GocParam{
		Service: service,
		Address: addr,
	}
	body, err := json.Marshal(&param)
	if err != nil {
		return "", fmt.Errorf("DeleteRegisterService json marshal failed: %w", err)
	}

	url := goc.host + CoverServicesRemoveAPI
	resp, err := goc.http.Post(ctx, url, getDefaultHeader(), string(body))
	if err != nil {
		return "", fmt.Errorf("DeleteRegisterService send http post failed: %w", err)
	}
	return string(resp), nil
}

//
// Profile API
//

// GetServiceProfileByName .
func (goc *GocAPI) GetServiceProfileByName(ctx context.Context, service string) ([]byte, error) {
	return goc.getServiceProfile(ctx, []string{service}, nil)
}

// GetServiceProfileByAddr .
func (goc *GocAPI) GetServiceProfileByAddr(ctx context.Context, addr string) ([]byte, error) {
	return goc.getServiceProfile(ctx, nil, []string{addr})
}

func (goc *GocAPI) getServiceProfile(ctx context.Context, service, addr []string) ([]byte, error) {
	param := GocParam{
		Service: service,
		Address: addr,
	}
	body, err := json.Marshal(&param)
	if err != nil {
		return nil, fmt.Errorf("getServiceProfile json marshal failed: %w", err)
	}

	url := goc.host + CoverProfileAPI
	resp, respProfile, err := goc.http.PostV2(ctx, url, getDefaultHeader(), string(body))
	if err != nil {
		return nil, fmt.Errorf("getServiceProfile send http post failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("getServiceProfile get non-200 returned code: %d, message: %s", resp.StatusCode, respProfile)
	}
	return respProfile, nil
}

// ClearProfileServiceByName .
func (goc *GocAPI) ClearProfileServiceByName(ctx context.Context, service string) (string, error) {
	return goc.clearServiceProfile(ctx, []string{service}, nil)
}

// ClearServiceProfileByAddr .
func (goc *GocAPI) ClearServiceProfileByAddr(ctx context.Context, addr string) (string, error) {
	return goc.clearServiceProfile(ctx, nil, []string{addr})
}

func (goc *GocAPI) clearServiceProfile(ctx context.Context, service, addr []string) (string, error) {
	param := GocParam{
		Service: service,
		Address: addr,
	}
	body, err := json.Marshal(&param)
	if err != nil {
		return "", fmt.Errorf("ClearProfile json marshal failed: %w", err)
	}

	url := goc.host + CoverProfileClearAPI
	resp, err := goc.http.Post(ctx, url, getDefaultHeader(), string(body))
	if err != nil {
		return "", fmt.Errorf("ClearProfile send http post failed: %w", err)
	}
	return string(resp), nil
}

func getDefaultHeader() map[string]string {
	return map[string]string{
		"Content-Type": "application/json",
	}
}

//
// Attach Server API
//

// APIGetServiceCoverage .
func APIGetServiceCoverage(ctx context.Context, host string) (string, error) {
	const coverageAPI = "/v1/cover/coverage"
	url := host + coverageAPI
	httpClient := utils.NewDefaultHTTPUtils()
	resp, respBody, err := httpClient.GetV2(ctx, url, map[string]string{})
	if err != nil {
		return "", fmt.Errorf("GetServiceCoverage send http get failed: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("GetAttachServiceCoverage get non-200 returned code: %d", resp.StatusCode)
	}
	return string(respBody), nil
}
