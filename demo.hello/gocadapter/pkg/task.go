package pkg

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"demo.hello/utils"
)

//
// Task remove unhealth services at fixed interval.
//

// TaskRemoveUnhealthServices .
func TaskRemoveUnhealthServices(ctx context.Context, interval time.Duration, host string) {
	go func() {
		tick := time.Tick(interval)
		for {
			select {
			case <-tick:
				func() {
					localCtx, cancel := context.WithTimeout(context.Background(), Wait)
					defer cancel()
					if err := removeUnhealthServicesFromGocSvrList(host); err != nil {
						errText := fmt.Sprintln("TaskRemoveUnhealthServices remove unhealth service failed:", err)
						notify.SendMessageToDefaultUser(localCtx, errText)
					}
				}()
			case <-ctx.Done():
				log.Println("Task exit: remove unhealth services from goc register list")
				return
			}
		}
	}()
}

// removeUnhealthServicesFromGocSvrList removes unhealth service from goc register services list.
func removeUnhealthServicesFromGocSvrList(host string) error {
	goc := NewGocAPI(host)
	ctx, cancel := context.WithTimeout(context.Background(), ShortWait)
	defer cancel()
	services, err := goc.ListRegisterServices(ctx)
	if err != nil {
		return fmt.Errorf("RemoveUnhealthServicesFromGocSvrList get goc register service list failed: %w", err)
	}

	for _, addrs := range services {
		for _, addr := range addrs {
			if err := func() error {
				localCtx, cancel := context.WithTimeout(context.Background(), LongWait)
				defer cancel()
				if !IsAttachServerOK(localCtx, addr) {
					if _, err := goc.DeleteRegisterServiceByAddr(ctx, addr); err != nil {
						return fmt.Errorf("RemoveUnhealthServicesFromGocSvrList remove goc register service [%s] failed: %w", addr, err)
					}
				}
				return nil
			}(); err != nil {
				return err
			}
		}
	}
	return nil
}

// IsAttachServerOK checks whether attach server is ok, and retry 3 times default.
func IsAttachServerOK(ctx context.Context, addr string) bool {
	for i := 1; ; i++ {
		if _, err := APIGetServiceCoverage(ctx, addr); err != nil {
			if i >= 4 {
				return false
			}
			log.Printf("IsAttachServerOK get service [%s] coverage failed: %s, retry %d", addr, err, i)
			time.Sleep(time.Duration(i) * time.Second)
		} else {
			return true
		}
	}
}

//
// Task sync service coverage/profile from goc center.
//

// SyncSrvCoverParam .
type SyncSrvCoverParam struct {
	Host      string
	Address   string
	SavedPath string
}

// SyncSrvCoverTaskManager .
type SyncSrvCoverTaskManager struct {
	tasks     map[string]*time.Timer
	intervals []time.Duration
	index     map[string]int
}

// NewSyncSrvCoverTaskManager .
func NewSyncSrvCoverTaskManager(intervals []time.Duration) *SyncSrvCoverTaskManager {
	return &SyncSrvCoverTaskManager{
		tasks:     make(map[string]*time.Timer, 8),
		intervals: intervals,
	}
}

// Exec .
func (m *SyncSrvCoverTaskManager) Exec(name string, param SyncSrvCoverParam) {
	m.index[name] = 0
	m.tasks[name] = time.AfterFunc(time.Second, func() {
		// TODO:
		idx := m.index[name]
		if idx >= len(m.intervals) {
			idx = len(m.intervals) - 1
		}
		m.tasks[name].Reset(m.intervals[idx])
		m.index[name] = idx + 1

	})
}

func getSrvCoverProcess(param SyncSrvCoverParam) (bool, error) {
	dirPath := filepath.Dir(param.SavedPath)
	coverFiles, err := utils.GetFilesBySuffix(dirPath, ".cov")
	if err != nil {
		return false, fmt.Errorf("getSrvCoverProcess get exsting profile file failed: %w", err)
	}
	if len(coverFiles) != 1 {
		return false, fmt.Errorf("getSrvCoverProcess get more than one existing profile file from dir: %s", dirPath)
	}

	if err := saveSrvCoverFile(param); err != nil {
		return false, fmt.Errorf("getSrvCoverProcess save service profile file failed: %w", err)
	}

	coverFilePath := coverFiles[0]
	coverFileName := filepath.Base(coverFilePath)
	historyCoverFilePath := filepath.Join(dirPath, "history", coverFileName)
	if err := utils.MoveFile(coverFilePath, historyCoverFilePath); err != nil {
		return false, fmt.Errorf("getSrvCoverProcess move exiting profile file to history failed: %w", err)
	}

	res, err := utils.IsFileContentEqual(historyCoverFilePath, param.SavedPath)
	if err != nil {
		return false, fmt.Errorf("getSrvCoverProcess compare profile files content failed: %w", err)
	}
	return res, nil
}

func saveSrvCoverFile(param SyncSrvCoverParam) error {
	ctx, cancel := context.WithTimeout(context.Background(), LongWait)
	defer cancel()

	goc := NewGocAPI(param.Host)
	b, err := goc.GetServiceProfileByAddr(ctx, param.Address)
	if err != nil {
		return fmt.Errorf("saveSrvCoverFile get service [%s] profile failed: %w", param.Address, err)
	}

	f, err := os.Create(param.SavedPath)
	if err != nil {
		return fmt.Errorf("saveSrvCoverFile open file [%s] failed: %w", param.SavedPath, err)
	}
	defer f.Close()

	buf := bytes.NewBuffer(b)
	if _, err := f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("saveSrvCoverFile write file [%s] failed: %w", param.SavedPath, err)
	}
	return nil
}

//
// Task delete history profile files.
//

// TODO:
