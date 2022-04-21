package pkg

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
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
						if err = s.notify.SendMessageToDefaultUser(localCtx, errText); err != nil {
							log.Println("RemoveUnhealthSrvTask send notify error:", err)
						}
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
				log.Println("SyncSrvCoverAndCreateReportTask run mock.")
			case <-ctx.Done():
				log.Println("SyncSrvCoverAndCreateReportTask exit.")
				return
			}
		}
	}()
}

func fetchAndSaveCoverForRegisterSrvs() {
	// TODO:
}
