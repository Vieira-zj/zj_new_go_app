package pkg

import (
	"log"
	"strings"
	"sync"
	"time"
)

/* Global state for srv cover sync task */

const (
	// StateRunning .
	StateRunning = iota
	// StateFreshed .
	StateFreshed
)

var (
	tasksState     *SrvCoverSyncTasksState
	tasksStateOnce sync.Once
)

// SrvCoverSyncTasksState .
type SrvCoverSyncTasksState struct {
	store map[string]int
	lock  *sync.RWMutex
}

// NewSrvCoverSyncTasksState .
func NewSrvCoverSyncTasksState() *SrvCoverSyncTasksState {
	tasksStateOnce.Do(func() {
		const defaultSize = 16
		tasksState = &SrvCoverSyncTasksState{
			store: make(map[string]int, defaultSize),
			lock:  &sync.RWMutex{},
		}
	})
	return tasksState
}

// Put .
func (state *SrvCoverSyncTasksState) Put(srvName string, srvState int) {
	defaultExpired := 15 * time.Minute
	state.PutByExpired(srvName, srvState, defaultExpired)
}

// PutByExpired .
func (state *SrvCoverSyncTasksState) PutByExpired(srvName string, srvState int, expired time.Duration) {
	state.lock.Lock()
	defer state.lock.Unlock()
	state.store[srvName] = srvState
	time.AfterFunc(expired, func() {
		delete(state.store, srvName)
	})
}

// Get .
func (state *SrvCoverSyncTasksState) Get(srvName string) (int, bool) {
	state.lock.RLock()
	defer state.lock.RUnlock()
	ret, ok := state.store[srvName]
	return ret, ok
}

// Delete .
func (state *SrvCoverSyncTasksState) Delete(srvName string) {
	state.lock.Lock()
	defer state.lock.Unlock()
	delete(state.store, srvName)
}

// Usage .
func (state *SrvCoverSyncTasksState) Usage() {
	usage := make(map[int][]string, 3)
	defaultSize := len(state.store) / 2
	for k, v := range state.store {
		if _, ok := usage[v]; !ok {
			usage[v] = make([]string, 0, defaultSize)
		}
		usage[v] = append(usage[v], k)
	}

	log.Println("Srv cover sync tasks state:")
	log.Printf("[Running]: count=%d, srv_keys={%s}", len(usage[StateRunning]), strings.Join(usage[StateRunning], ","))
	log.Printf("[Refreshed]: count=%d, srv_keys={%s}", len(usage[StateFreshed]), strings.Join(usage[StateFreshed], ","))
}
