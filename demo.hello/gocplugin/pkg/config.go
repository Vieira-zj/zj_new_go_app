package pkg

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

/* Config */

const (
	shortWait = 3 * time.Second
	wait      = 5 * time.Second
	longWait  = 8 * time.Second
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

/* Srv Cover Sync Task State */

const (
	// StateRunning .
	StateRunning = iota
	// StateFreshed .
	StateFreshed
	// StateExpired .
	StateExpired
)

// SrvCoverSyncTasksState .
type SrvCoverSyncTasksState struct {
	store map[string]bool
	lock  *sync.RWMutex
}

// NewSrvCoverSyncTasksState .
func NewSrvCoverSyncTasksState() *SrvCoverSyncTasksState {
	// TODO:
	return nil
}

// Set .
func (state *SrvCoverSyncTasksState) Set(key string, value int, expired time.Duration) {
	time.AfterFunc(expired, func() {})
}

// Get .
func (state *SrvCoverSyncTasksState) Get(key string) (int, error) {
	state.lock.Lock()
	defer state.lock.Unlock()
	return 0, nil
}

// Usage .
func (state *SrvCoverSyncTasksState) Usage() {}
