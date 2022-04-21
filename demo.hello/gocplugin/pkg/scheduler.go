package pkg

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"demo.hello/utils"
)

var (
	scheduler     *Scheduler
	schedulerOnce sync.Once
)

// Scheduler runs schedule tasks.
type Scheduler struct {
	notify *MatterMostNotify
}

// NewScheduler .
func NewScheduler() *Scheduler {
	schedulerOnce.Do(func() {
		scheduler = &Scheduler{
			notify: NewMatterMostNotify(),
		}
	})
	return scheduler
}

// RemoveUnhealthSrvTask .
func (s *Scheduler) RemoveUnhealthSrvTask(ctx context.Context, interval time.Duration) {
	go func() {
		tick := time.Tick(interval)
		for {
			select {
			case <-tick:
				func() {
					if err := RemoveUnhealthSrvInGocTask(); err != nil {
						localCtx, cancel := context.WithTimeout(context.Background(), Wait)
						defer cancel()
						errText := fmt.Sprintln("RemoveUnhealthSrvTask error:", err)
						s.notify.MustSendMessageToDefaultUser(localCtx, errText)
					}
				}()
			case <-ctx.Done():
				log.Println("RemoveUnhealthSrvTask exit.")
				return
			}
		}
	}()
}

// SyncRegisterSrvsCoverTask .
func (s *Scheduler) SyncRegisterSrvsCoverTask(ctx context.Context, interval time.Duration) {
	go func() {
		tick := time.Tick(interval)
		for {
			select {
			case <-tick:
				func() {
					if err := fetchAndSaveCoverForRegisterSrvs(); err != nil {
						localCtx, cancel := context.WithTimeout(context.Background(), Wait)
						defer cancel()
						errText := fmt.Sprintln("SyncRegisterSrvsCoverTask error:", err)
						s.notify.MustSendMessageToDefaultUser(localCtx, errText)
					}
				}()
			case <-ctx.Done():
				log.Println("SyncRegisterSrvsCoverTask exit.")
				return
			}
		}
	}()
}

func fetchAndSaveCoverForRegisterSrvs() error {
	ctx, cancel := context.WithTimeout(context.Background(), Wait)
	defer cancel()

	gocAPI := NewGocAPI()
	srvs, err := gocAPI.ListRegisterServices(ctx)
	if err != nil {
		return fmt.Errorf("fetchAndSaveCoverForRegisterSrvs error: %w", err)
	}

	for srvName, addrs := range srvs {
		dir := GetModuleDir(srvName)
		savedDir := filepath.Join(dir, WatcherCoverDataDirName)
		if utils.IsDirExist(savedDir) {
			utils.MakeDir(savedDir)
		}

		lastCovFilePath, err := utils.GetLatestFileInDir(savedDir, "cov")
		if err != nil && !errors.Is(err, utils.ErrNoFilesExistInDir) {
			return fmt.Errorf("fetchAndSaveCoverForRegisterSrvs error: %w", err)
		}

		param := SyncSrvCoverParam{
			SrvName:   srvName,
			Addresses: addrs,
		}
		savedCovPath, err := FetchAndSaveSrvCover(savedDir, param)
		if err != nil {
			return fmt.Errorf("fetchAndSaveCoverForRegisterSrvs error: %w", err)
		}
		removeLastEqualCovFile(lastCovFilePath, savedCovPath)
	}
	return nil
}

func removeLastEqualCovFile(lastCovFilePath, curCovFilePath string) {
	if len(lastCovFilePath) == 0 {
		return
	}

	isEqual, err := utils.IsFilesEqual(lastCovFilePath, curCovFilePath)
	if err != nil {
		log.Println("fetchAndSaveCoverForRegisterSrvs file compare error:", err)
		return
	}
	if isEqual {
		if err := os.Remove(lastCovFilePath); err != nil {
			log.Println("fetchAndSaveCoverForRegisterSrvs remove file error:", err)
		}
	}
}
