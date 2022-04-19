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
		srvCoverSyncTasksPool = utils.NewGoPool(10, 100, 15*time.Second)
	})
	srvCoverSyncTasksPool.Start()
}

// CloseSrvCoverSyncTasksPool .
func CloseSrvCoverSyncTasksPool() {
	if srvCoverSyncTasksPool != nil {
		const stopTimeout = 60
		srvCoverSyncTasksPool.Stop(stopTimeout)
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
			retCh <- fmt.Errorf("Asyc run GetSrvCoverAndCreateReportTask error: %w", err)
		} else {
			tasksState.Put(param.SrvName, StateFreshed)
			retCh <- coverTotal
		}
	}, submitTimeout); err != nil {
		retCh <- fmt.Errorf("Submit GetSrvCoverAndCreateReportTask error: %w", err)
	}
	return retCh
}
