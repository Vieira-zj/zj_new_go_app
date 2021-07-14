package handlers

import (
	"context"
	"fmt"
	"time"
)

var results []*JobResult

func init() {
	buildMockJobsResults(10)
}

func buildMockJobsResults(count int) {
	results = make([]*JobResult, 0, count)
	for i := 0; i < count; i++ {
		result := &JobResult{
			ID:     uint32(i),
			Title:  fmt.Sprintf("Job:%d", i),
			Status: getJobStatus(i),
		}
		results = append(results, result)
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
	return results
}

func getMockDeltaJobsResults(ctx context.Context) chan *JobResult {
	ch := make(chan *JobResult, 10)
	go func() {
		for {
			notDoneJobResults := getNotDoneJobResults()
			if len(notDoneJobResults) == 0 {
				close(ch)
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
					ch <- result
				}
			}
		}
	}()
	return ch
}

func getNotDoneJobResults() []*JobResult {
	ret := make([]*JobResult, 0)
	for _, result := range results {
		if result.Status != "done" {
			ret = append(ret, result)
		}
	}
	return ret
}
