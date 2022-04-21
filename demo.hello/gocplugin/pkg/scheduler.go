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
						errMsg := fmt.Sprintln("RemoveUnhealthSrvTask error:", err)
						s.notify.MustSendMessageToDefaultUser(errMsg)
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
				if err := s.fetchAndSaveCoverForRegisterSrvs(); err != nil {
					errMsg := fmt.Sprintln("SyncRegisterSrvsCoverTask error:", err)
					s.notify.MustSendMessageToDefaultUser(errMsg)
				}
			case <-ctx.Done():
				log.Println("SyncRegisterSrvsCoverTask exit.")
				return
			}
		}
	}()
}

func (s *Scheduler) fetchAndSaveCoverForRegisterSrvs() error {
	log.Println("fetchAndSaveCoverForRegisterSrvs auto trigger")
	ctx, cancel := context.WithTimeout(context.Background(), Wait)
	defer cancel()

	gocAPI := NewGocAPI()
	srvs, err := gocAPI.ListRegisterServices(ctx)
	if err != nil {
		return fmt.Errorf("fetchAndSaveCoverForRegisterSrvs error: %w", err)
	}

	const limit = 5
	semaphore := make(chan struct{}, limit)
	var wg sync.WaitGroup

	for srvName := range srvs {
		wg.Add(1)
		go func(srvName string) {
			semaphore <- struct{}{}
			defer func() {
				<-semaphore
				wg.Done()
			}()

			savedDir := filepath.Join(GetModuleDir(srvName), WatcherCoverDataDirName)
			if !utils.IsDirExist(savedDir) {
				utils.MakeDir(savedDir)
			}

			lastCovFileName, err := utils.GetLatestFileInDir(savedDir, "cov")
			if err != nil && !errors.Is(err, utils.ErrNoFilesExistInDir) {
				errMsg := fmt.Sprintln("fetchAndSaveCoverForRegisterSrvs error:", err)
				s.notify.MustSendMessageToDefaultUser(errMsg)
				return
			}

			savedCovPath, err := FetchAndSaveSrvCover(savedDir, srvName)
			if err != nil {
				errMsg := fmt.Sprintln("fetchAndSaveCoverForRegisterSrvs error:", err)
				s.notify.MustSendMessageToDefaultUser(errMsg)
				return
			}
			removeLastEqualCovFile(filepath.Join(savedDir, lastCovFileName), savedCovPath)
		}(srvName)
	}
	wg.Wait()

	return nil
}

// RemoveExpiredSrvCoverFilesTask .
func (s *Scheduler) RemoveExpiredSrvCoverFilesTask() {
	// TODO:
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
