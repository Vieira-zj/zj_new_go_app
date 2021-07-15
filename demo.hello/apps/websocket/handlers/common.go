package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"demo.hello/utils"
)

var (
	// EventBus .
	EventBus   *utils.EventBusServer
	jobResults *JobResults
	channel    string
)

func init() {
	channel = "OnJobResultsUpdateStatus"
	EventBus = utils.NewEventBusServer(10)
	EventBus.Start()
	buildMockJobResults(10)
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
type JobResults struct {
	items []*JobResult
}

// UpdateStatus .
func (results *JobResults) UpdateStatus(result *JobResult) error {
	return EventBus.Publish(channel, result)
}

/*
create mock data
*/

func buildMockJobResults(count int) {
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
		items: results,
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

func getMockJobResults() []*JobResult {
	return jobResults.items
}

/*
create mock delta data
*/

func getMockDeltaJobResults(ctx context.Context) (err error) {
	fmt.Println("start update job")
	for {
		notDoneJobResults := getNotDoneJobResults()
		if len(notDoneJobResults) == 0 {
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
	for _, result := range jobResults.items {
		if result.Status != "done" {
			ret = append(ret, result)
		}
	}
	return ret
}
