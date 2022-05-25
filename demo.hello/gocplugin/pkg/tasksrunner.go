package pkg

import (
	"fmt"
	"sync"
	"time"

	"demo.hello/utils"
)

var (
	srvCoverSyncTasksPool     *utils.GoPool
	srvCoverSyncTasksPoolOnce sync.Once
)

// InitSrvCoverSyncTasksPool .
func InitSrvCoverSyncTasksPool() {
	srvCoverSyncTasksPoolOnce.Do(func() {
		const (
			coreSize  = 10
			queueSize = 100
			idleTime  = 3 * time.Minute
		)
		srvCoverSyncTasksPool = utils.NewGoPool(coreSize, queueSize, idleTime)
	})
	srvCoverSyncTasksPool.Start()
}

// CloseSrvCoverSyncTasksPool .
func CloseSrvCoverSyncTasksPool() {
	if srvCoverSyncTasksPool != nil {
		const stopWaitSec = 60
		srvCoverSyncTasksPool.Stop(stopWaitSec)
	}
}

// SubmitSrvCoverSyncTask .
func SubmitSrvCoverSyncTask(param SyncSrvCoverParam) chan interface{} {
	const submitTimeout = 3 * time.Minute
	retCh := make(chan interface{}, 1)

	if err := srvCoverSyncTasksPool.SubmitWithTimeout(func() {
		tasksState := NewSrvCoverSyncTasksState()
		tasksState.Put(param.SrvName, StateRunning)
		if coverTotal, err := GetSrvCoverAndCreateReportTask(param); err != nil {
			tasksState.Delete(param.SrvName)
			notify := NewMatterMostNotify()
			err := fmt.Errorf("Async run GetSrvCoverAndCreateReportTask error: %w", err)
			notify.MustSendMessageToDefaultUser(err.Error())
			retCh <- err
		} else {
			tasksState.Put(param.SrvName, StateFreshed)
			retCh <- coverTotal
		}
	}, submitTimeout); err != nil {
		retCh <- fmt.Errorf("Submit GetSrvCoverAndCreateReportTask error: %w", err)
	}

	return retCh
}
