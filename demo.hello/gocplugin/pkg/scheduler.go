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
				if err := RemoveUnhealthSrvInGocTask(); err != nil {
					errMsg := fmt.Sprintln("RemoveUnhealthSrvTask error:", err)
					s.notify.MustSendMessageToDefaultUser(errMsg)
				}
			case <-ctx.Done():
				log.Println("RemoveUnhealthSrvTask exit.")
				return
			}
		}
	}()
}

// SyncRegisterSrvsCoverReportTask Sync service cover file and create report. (trigger from report)
func (s *Scheduler) SyncRegisterSrvsCoverReportTask(ctx context.Context, interval time.Duration) {
	go func() {
		tick := time.Tick(interval)
		for {
			select {
			case <-tick:
				if err := syncRegisterSrvsCoverReport(); err != nil {
					errMsg := fmt.Sprintln("SyncRegisterSrvsCoverReportTask error:", err)
					s.notify.MustSendMessageToDefaultUser(errMsg)
				}
			case <-ctx.Done():
				log.Println("SyncRegisterSrvsCoverReportTask exit.")
				return
			}
		}
	}()
}

// SyncSrvsRawCoverTask Sync service raw ".cov" cover file. (trigger from watcher)
func (s *Scheduler) SyncSrvsRawCoverTask(ctx context.Context, interval time.Duration) {
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
	log.Println("[fetchAndSaveCoverForRegisterSrvs] start")
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

			savedDir := filepath.Join(GetSrvModuleDir(srvName), WatcherCoverDataDirName)
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
			removeDuplicatedCovFile(filepath.Join(savedDir, lastCovFileName), savedCovPath)
		}(srvName)
	}
	wg.Wait()

	return nil
}

func removeDuplicatedCovFile(preCovFilePath, curCovFilePath string) {
	if len(preCovFilePath) == 0 {
		return
	}

	isEqual, err := utils.IsFilesEqual(preCovFilePath, curCovFilePath)
	if err != nil {
		log.Println("removeDuplicatedCovFile file compare error:", err)
		return
	}

	if isEqual {
		if err := os.Remove(preCovFilePath); err != nil {
			log.Println("removeDuplicatedCovFile remove file error:", err)
		}
	}
}

// RemoveExpiredSrvCoverFilesTask .
func (s *Scheduler) RemoveExpiredSrvCoverFilesTask() {
	// TODO:
}

//
// Common
//

func syncRegisterSrvsCoverReport() error {
	log.Println("[syncRegisterSrvsCoverReport] start")
	srvs, err := SyncAndListRegisterSrvsTask()
	if err != nil {
		return fmt.Errorf("syncRegisterSrvsCoverReport error: %w", err)
	}

	tasksState := NewSrvCoverSyncTasksState()
	for srv := range srvs {
		// fitler running or refreshed tasks
		if _, ok := tasksState.Get(srv); ok {
			delete(srvs, srv)
		}
	}

	var wg sync.WaitGroup
	for srv, addrs := range srvs {
		wg.Add(1)
		go func(srv string, addrs []string) {
			retCh := SubmitSrvCoverSyncTask(SyncSrvCoverParam{
				SrvName:   srv,
				Addresses: addrs,
			})
			<-retCh
			close(retCh)
			wg.Done()
		}(srv, addrs)
	}

	wg.Wait()
	log.Println("[syncRegisterSrvsCoverReport] end")
	return nil
}
