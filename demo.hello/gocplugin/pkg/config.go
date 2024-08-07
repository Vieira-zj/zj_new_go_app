package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	// ShortWait .
	ShortWait = 3 * time.Second
	// Wait .
	Wait = 5 * time.Second
	// LongWait .
	LongWait = 10 * time.Second

	// ReportCoverDataDirName .
	ReportCoverDataDirName = "cover_data"
	// WatcherCoverDataDirName .
	WatcherCoverDataDirName = "history_cov_files"

	// CoverRptTypeRaw .
	CoverRptTypeRaw = "cov"
	// CoverRptTypeFunc .
	CoverRptTypeFunc = "func"
	// CoverRptTypeHTML .
	CoverRptTypeHTML = "html"

	// RunModeReport .
	RunModeReport = "report"
	// RunModeWatcher .
	RunModeWatcher = "watcher"
)

var (
	// AppConfig .
	AppConfig GocPluginConfig
	// ModuleToRepoMap .
	ModuleToRepoMap map[string]string
)

// GocPluginConfig .
type GocPluginConfig struct {
	RunMode          string `json:"run_mode"`
	IsDebug          bool   `json:"is_debug"`
	RootDir          string `json:"root"`
	PublicDir        string
	GocCenterIngHost string `json:"goc_center_ing_host"`
	GocCenterSvcHost string `json:"goc_center_svc_host"`
	PodMonitorHost   string `json:"pod_monitor_host"`
}

// InitConfig .
func InitConfig(rootDir string) error {
	AppConfig.RootDir = rootDir
	const cfgFileName = "gocplugin.json"
	cfgPath := filepath.Join(AppConfig.RootDir, cfgFileName)
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("InitConfig read config file error: %w", err)
	}

	if err := json.Unmarshal(b, &AppConfig); err != nil {
		return fmt.Errorf("InitConfig error: %w", err)
	}
	AppConfig.PublicDir = filepath.Join(AppConfig.RootDir, "public/report")

	err = LoadModuleToRepoMap()
	return err
}

// LoadModuleToRepoMap .
func LoadModuleToRepoMap() error {
	const mapFile = "module_repo_map.json"
	b, err := os.ReadFile(filepath.Join(AppConfig.RootDir, mapFile))
	if err != nil {
		return fmt.Errorf("LoadModuleToRepoMap error: %w", err)
	}

	if err := json.Unmarshal(b, &ModuleToRepoMap); err != nil {
		return fmt.Errorf("LoadModuleToRepoMap error: %w", err)
	}
	return nil
}

// GetModuleCoverDataDir .
func GetModuleCoverDataDir(appName string) string {
	return filepath.Join(AppConfig.RootDir, appName, ReportCoverDataDirName)
}

func getGocCenterHost() string {
	host := AppConfig.GocCenterIngHost
	if AppConfig.RunMode == RunModeWatcher {
		host = AppConfig.GocCenterSvcHost
	}
	return host
}
