package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const (
	shortWait = 3 * time.Second
	wait      = 5 * time.Second
	longWait  = 8 * time.Second
)

var (
	// AppConfig .
	AppConfig AdapterConfig
	// ModuleToRepoMap .
	ModuleToRepoMap map[string]string
)

// AdapterConfig .
type AdapterConfig struct {
	RootDir string `json:"root"`
	GocHost string `json:"host"`
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
	return nil
}

// LoadModuleToRepoMap .
func LoadModuleToRepoMap() error {
	jsonFile := "module_repo_map.json"
	b, err := os.ReadFile(filepath.Join(AppConfig.RootDir, jsonFile))
	if err != nil {
		return fmt.Errorf("LoadModuleToRepoMap error: %w", err)
	}

	if err := json.Unmarshal(b, &ModuleToRepoMap); err != nil {
		return fmt.Errorf("LoadModuleToRepoMap error: %w", err)
	}
	return nil
}