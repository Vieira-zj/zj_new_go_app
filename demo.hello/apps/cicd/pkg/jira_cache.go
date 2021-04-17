package pkg

import (
	"context"
	"fmt"

	"demo.hello/utils"
)

/*
Worker
*/

// JiraTaskWorker .
type JiraTaskWorker struct {
	parallel int
	ctx      context.Context
	queue    chan string
	jira     *JiraTool
	Cache    *JiraPMTaskCache
}

// NewJiraTaskWorker .
func NewJiraTaskWorker(ctx context.Context, parallel int) *JiraTaskWorker {
	return &JiraTaskWorker{
		parallel: parallel,
		ctx:      ctx,
		queue:    make(chan string, (parallel + 1)),
		jira:     NewJiraTool(),
		Cache:    NewJiraPMTaskCache(parallel, 20),
	}
}

// QueueSize .
func (worker *JiraTaskWorker) QueueSize() int {
	return len(worker.queue)
}

// Submit .
func (worker *JiraTaskWorker) Submit(issueID string) {
	worker.queue <- issueID
}

// Run .
func (worker *JiraTaskWorker) Run() {
	for i := 0; i < worker.parallel; i++ {
		go func() {
			var issueID string
			for {
				select {
				case issueID = <-worker.queue:
					fmt.Println("Work on issue:", issueID)
				case <-worker.ctx.Done():
					fmt.Println("Worker exit.")
					return
				}

				issue, err := NewJiraIssue(context.TODO(), worker.jira, issueID)
				if err != nil {
					fmt.Println("Create jira issue failed:", err)
				}
				worker.Cache.PutTask(issueID, issue, issue.Type)
				if issue.Type == "PMTask" {
					for _, subIssueID := range issue.SubIssues {
						worker.queue <- subIssueID
					}
				}
			}
		}()
	}
}

/*
Cache
*/

// JiraPMTaskCache .
type JiraPMTaskCache struct {
	PMTasks *utils.Cache
	Tasks   *utils.Cache
}

// NewJiraPMTaskCache .
func NewJiraPMTaskCache(shardNumber, mapSize int) *JiraPMTaskCache {
	return &JiraPMTaskCache{
		PMTasks: utils.NewCache(shardNumber, mapSize),
		Tasks:   utils.NewCache((shardNumber + 3), mapSize),
	}
}

// GetTask .
func (cache *JiraPMTaskCache) GetTask(key string, taskType string) (interface{}, error) {
	if taskType == "PMTask" {
		return cache.PMTasks.Get(key)
	}
	return cache.Tasks.Get(key)
}

// PutTask .
func (cache *JiraPMTaskCache) PutTask(key string, value *JiraIssue, taskType string) {
	if taskType == "PMTask" {
		cache.PMTasks.Put(key, value)
	} else {
		cache.Tasks.Put(key, value)
	}
}

// PrintTaskTree .
func (cache *JiraPMTaskCache) PrintTaskTree() {
	fmt.Println("[PM Tasks:]")
	tasks := cache.PMTasks.GetItems()
	subTaskKeys := make([]string, 10)
	for _, task := range tasks {
		issue := task.(*JiraIssue)
		issue.PrintText("", false)
		subTaskKeys = append(subTaskKeys, issue.SubIssues...)
		for _, key := range issue.SubIssues {
			subIssue, err := cache.Tasks.Get(key)
			if err != nil {
				fmt.Println("Get sub issue failed:", err)
			}
			subIssue.(*JiraIssue).PrintText("\t", false)
		}
		fmt.Println()
	}

	fmt.Println("\n[Tasks:]")
	tasks = cache.Tasks.GetItems()
	for key, task := range tasks {
		isSubTask := false
		for _, subTask := range subTaskKeys {
			if key == subTask {
				isSubTask = true
				break
			}
		}
		if !isSubTask {
			issue := task.(*JiraIssue)
			issue.PrintText("", false)
		}
	}
}
