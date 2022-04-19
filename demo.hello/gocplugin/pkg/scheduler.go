package pkg

import (
	"context"
	"fmt"
	"log"
	"time"
)

/* Run scheduled tasks */

func scheduleTaskRemoveUnhealthSrv(ctx context.Context, interval time.Duration) {
	go func() {
		tick := time.Tick(interval)
		for {
			select {
			case <-tick:
				func() {
					if err := removeUnhealthSrvInGocTask(); err != nil {
						localCtx, cancel := context.WithTimeout(context.Background(), Wait)
						defer cancel()
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

func scheduleTaskSyncSrvCoverAndCreateReport(param SyncSrvCoverParam, intervals []time.Duration) error {
	// TODO:
	return nil
}
