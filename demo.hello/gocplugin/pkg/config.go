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
	LongWait = 8 * time.Second

	// ReportCoverDataDirName .
	ReportCoverDataDirName = "cover_data"
	// WatcherCoverDataDirName .
	WatcherCoverDataDirName = "saved_cover_data"
)

var (
	// AppConfig .
	AppConfig GocPluginConfig
	// ModuleToRepoMap .
	ModuleToRepoMap map[string]string
)

// GocPluginConfig .
type GocPluginConfig struct {
	RootDir string `json:"root"`
	GocHost string `json:"goc_host"`
}

// LoadConfig .
func LoadConfig(cfgPath string) error {
	b, err := os.ReadFile(cfgPath)
	if err != nil {
		return fmt.Errorf("LoadConfig read config file error: %w", err)
	}

	if err := json.Unmarshal(b, &AppConfig); err != nil {
		return fmt.Errorf("LoadConfig error: %w", err)
	}

	LoadModuleToRepoMap()
	return nil
}

// LoadModuleToRepoMap .
func LoadModuleToRepoMap() error {
	const jsonFile = "module_repo_map.json"
	b, err := os.ReadFile(filepath.Join(AppConfig.RootDir, jsonFile))
	if err != nil {
		return fmt.Errorf("LoadModuleToRepoMap error: %w", err)
	}

	if err := json.Unmarshal(b, &ModuleToRepoMap); err != nil {
		return fmt.Errorf("LoadModuleToRepoMap error: %w", err)
	}
	return nil
}