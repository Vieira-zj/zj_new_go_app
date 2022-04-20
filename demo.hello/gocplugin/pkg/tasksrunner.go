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
			idleTime  = time.Minute
			coreSize  = 10
			queueSize = 30
		)
		srvCoverSyncTasksPool = utils.NewGoPool(coreSize, coreSize+queueSize, idleTime)
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
func SubmitSrvCoverSyncTask(param SyncSrvCoverParam) <-chan interface{} {
	const submitTimeout = time.Minute
	retCh := make(chan interface{}, 1)

	if err := srvCoverSyncTasksPool.SubmitWithTimeout(func() {
		tasksState := NewSrvCoverSyncTasksState()
		tasksState.Put(param.SrvName, StateRunning)
		if coverTotal, err := GetSrvCoverAndCreateReportTask(param); err != nil {
			tasksState.Delete(param.SrvName)
			retCh <- fmt.Errorf("Async run GetSrvCoverAndCreateReportTask error: %w", err)
		} else {
			tasksState.Put(param.SrvName, StateFreshed)
			retCh <- coverTotal
		}
	}, submitTimeout); err != nil {
		retCh <- fmt.Errorf("Submit GetSrvCoverAndCreateReportTask error: %w", err)
	}
	return retCh
}
