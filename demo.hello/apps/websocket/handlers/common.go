package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"demo.hello/utils"
	"github.com/gorilla/websocket"
)

var (
	// EventBus .
	EventBus   *utils.EventBusServer
	client     *websocket.Upgrader
	jobResults JobResults
	mock       *Mock
	channel    string
)

func init() {
	client = &websocket.Upgrader{
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		HandshakeTimeout: time.Duration(3) * time.Second,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	mock = &Mock{
		locker: &sync.RWMutex{},
	}
	mock.buildJobResults(10)

	channel = "OnJobResultsUpdateStatus"
	EventBus = utils.NewEventBusServer(16, 0)
}

// ResponseData .
type ResponseData struct {
	Code    uint32          `json:"code"`
	Results json.RawMessage `json:"results"`
}

// JobResult .
type JobResult struct {
	ID     uint32 `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

// JobResults .
type JobResults []*JobResult

// UpdateStatus .
func (results *JobResults) UpdateStatus(result *JobResult) error {
	// only publish event for test
	return EventBus.Publish(channel, result)
}

/*
create mock data
*/

// Mock .
type Mock struct {
	isRunning bool
	locker    *sync.RWMutex
}

func (mock *Mock) buildJobResults(count int) {
	results := make([]*JobResult, 0, count)
	for i := 0; i < count; i++ {
		result := &JobResult{
			ID:     uint32(i),
			Title:  fmt.Sprintf("Job:%d", i),
			Status: getJobStatus(i),
		}
		results = append(results, result)
	}
	jobResults = results
}

func getJobStatus(num int) string {
	if num%3 == 0 {
		return "notstart"
	} else if num%4 == 0 {
		return "done"
	}
	return "running"
}

/*
create mock delta data
*/

func (mock *Mock) getDeltaJobResults(ctx context.Context) (err error) {
	// 1. 防止重复执行；2. 保证并发安全
	mock.locker.Lock()
	if mock.isRunning {
		mock.locker.Unlock()
		return
	}
	mock.isRunning = true
	mock.locker.Unlock()

	fmt.Println("start delta update process")
	for {
		notDoneJobResults := getNotDoneJobResults()
		if len(notDoneJobResults) == 0 {
			mock.isRunning = false
			return
		}

		for _, result := range notDoneJobResults {
			if result.Status == "running" {
				result.Status = "done"
			} else if result.Status == "notstart" {
				result.Status = "running"
			} else {
				fmt.Println("invalid job result:", result)
				continue
			}
			time.Sleep(time.Second)

			select {
			case <-ctx.Done():
				fmt.Println("timeout, and getMockDeltaJobsResults return")
				return
			default:
				if err = jobResults.UpdateStatus(result); err != nil {
					return
				}
			}
		}
	}
}

func getNotDoneJobResults() []*JobResult {
	ret := make([]*JobResult, 0)
	for _, result := range jobResults {
		if result.Status != "done" {
			ret = append(ret, result)
		}
	}
	return ret
}
