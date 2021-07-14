package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

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
type JobResults struct {
	items             []*JobResult
	OnUpdateStatusCbs map[string]func(*JobResult)
}

// RegisterCallback .
func (results *JobResults) RegisterCallback(key string, cb func(*JobResult)) {
	results.OnUpdateStatusCbs[key] = cb
}

// UnRegisterCallback .
func (results *JobResults) UnRegisterCallback(key string) {
	delete(results.OnUpdateStatusCbs, key)
}

// UpdateStatus .
func (results *JobResults) UpdateStatus(result *JobResult, status string) {
	result.Status = status
	for _, fn := range results.OnUpdateStatusCbs {
		// fmt.Println("run OnUpdateStatusCbs:", key)
		fn(result)
	}
}

/*
create mock data
*/

var jobResults *JobResults

func init() {
	buildMockJobsResults(10)
}

func buildMockJobsResults(count int) {
	results := make([]*JobResult, 0, count)
	for i := 0; i < count; i++ {
		result := &JobResult{
			ID:     uint32(i),
			Title:  fmt.Sprintf("Job:%d", i),
			Status: getJobStatus(i),
		}
		results = append(results, result)
	}
	jobResults = &JobResults{
		items:             results,
		OnUpdateStatusCbs: make(map[string]func(*JobResult), 0),
	}
}

func getJobStatus(num int) string {
	if num%3 == 0 {
		return "notstart"
	} else if num%4 == 0 {
		return "done"
	}
	return "running"
}

func getMockJobsResults() []*JobResult {
	return jobResults.items
}

/*
create mock delta data
*/

var locker = &sync.Mutex{}
var isRunning bool
var done chan struct{}

func getMockDeltaJobsResults(ctx context.Context) chan struct{} {
	locker.Lock()
	defer locker.Unlock()

	if isRunning {
		return done
	}

	isRunning = true
	done = make(chan struct{})
	go func() {
		fmt.Println("start update job")
		defer func() {
			isRunning = false
			close(done)
		}()
		for {
			notDoneJobResults := getNotDoneJobResults()
			if len(notDoneJobResults) == 0 {
				return
			}

			for _, result := range notDoneJobResults {
				newStatus := ""
				if result.Status == "running" {
					newStatus = "done"
				} else if result.Status == "notstart" {
					newStatus = "running"
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
					jobResults.UpdateStatus(result, newStatus)
				}
			}
		}
	}()
	return done
}

func getNotDoneJobResults() []*JobResult {
	ret := make([]*JobResult, 0)
	for _, result := range jobResults.items {
		if result.Status != "done" {
			ret = append(ret, result)
		}
	}
	return ret
}
