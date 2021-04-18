package pkg

import (
	"context"
	"fmt"
	"time"

	"demo.hello/utils"
)

/*
Worker
*/

// Jira2LevelsTreeWorker handle jira issues and save in cache for search.
type Jira2LevelsTreeWorker struct {
	Key      string
	parallel int
	ctx      context.Context
	queue    chan string
	jira     *JiraTool
	store    *Jira2LevelsTreeStore
}

// NewJira2LevelsTreeWorker create an instance of Jira2LevelsTreeWorker.
func NewJira2LevelsTreeWorker(ctx context.Context, key string, parallel int) *Jira2LevelsTreeWorker {
	// 1.queueSize设置太小可能会导致阻塞 2.分片存储，mapSize不需要设置过大
	const (
		queueSize = 30
		mapSize   = 20
	)
	return &Jira2LevelsTreeWorker{
		Key:      key,
		parallel: parallel,
		ctx:      ctx,
		queue:    make(chan string, queueSize),
		jira:     NewJiraTool(),
		store:    NewJira2LevelsTreeStore(parallel, mapSize),
	}
}

// QueueSize returns total issue keys to be handle in queue.
func (worker *Jira2LevelsTreeWorker) QueueSize() int {
	return len(worker.queue)
}

// GetStore returns internal store.
func (worker *Jira2LevelsTreeWorker) GetStore() *Jira2LevelsTreeStore {
	return worker.store
}

// Submit puts a jira issue key in queue.
func (worker *Jira2LevelsTreeWorker) Submit(issueID string) {
	worker.queue <- issueID
}

// Start run worker to handle jira issues and save in store.
func (worker *Jira2LevelsTreeWorker) Start() {
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

				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
				issue, err := NewJiraIssue(ctx, worker.jira, issueID)
				cancel()
				if err != nil {
					fmt.Println("Create jira issue failed:", err)
					continue
				}

				worker.store.PutTask(issueID, issue)
				if isLevel1Task(issue.Type) {
					for _, subIssueID := range issue.SubIssues {
						if !worker.store.IsTaskExist(subIssueID) {
							worker.queue <- subIssueID
						}
					}
				}
			}
		}()
	}
}

/*
Store
*/

// Jira2LevelsTreeStore stores pm->task and story->task tree.
type Jira2LevelsTreeStore struct {
	Level1Tasks *utils.Cache
	Level2Tasks *utils.Cache
}

// NewJira2LevelsTreeStore creates an instance of Jira2LevelsTreeStore.
func NewJira2LevelsTreeStore(shardNumber, mapSize int) *Jira2LevelsTreeStore {
	return &Jira2LevelsTreeStore{
		Level1Tasks: utils.NewCache(shardNumber, mapSize),
		Level2Tasks: utils.NewCache((shardNumber + 3), mapSize),
	}
}

// GetTask returns a task by type.
func (store *Jira2LevelsTreeStore) GetTask(key string, taskType string) (interface{}, error) {
	if isLevel1Task(taskType) {
		return store.Level1Tasks.Get(key)
	}
	return store.Level2Tasks.Get(key)
}

// PutTask puts a task by type.
func (store *Jira2LevelsTreeStore) PutTask(key string, issue *JiraIssue) {
	if isLevel1Task(issue.Type) {
		store.Level1Tasks.Put(key, issue)
	} else {
		store.Level2Tasks.Put(key, issue)
	}
}

// IsTaskExist returns whether task is exist in store.
func (store *Jira2LevelsTreeStore) IsTaskExist(key string) bool {
	if store.Level1Tasks.IsExist(key) || store.Level2Tasks.IsExist(key) {
		return true
	}
	return false
}

// PrintTree prints pm->task and story->task tree.
func (store *Jira2LevelsTreeStore) PrintTree() {
	var subTasksMap map[string]struct{}
	tasks := store.Level1Tasks.GetItems()
	if len(tasks) > 0 {
		subTasksMap = make(map[string]struct{}, 10)
		fmt.Println("\n[Tasks and Sub Tasks:]")
		for _, task := range tasks {
			issue := task.(*JiraIssue)
			issue.PrintText("")
			for _, key := range issue.SubIssues {
				subTasksMap[key] = struct{}{}
				if subIssue, err := store.Level2Tasks.Get(key); err != nil {
					fmt.Println("Get sub issue failed:", err)
				} else {
					subIssue.(*JiraIssue).PrintText("\t")
				}
			}
			fmt.Println()
		}
	}

	fmt.Println("\n[Tasks:]")
	tasks = store.Level2Tasks.GetItems()
	for key, task := range tasks {
		if _, ok := subTasksMap[key]; !ok {
			issue := task.(*JiraIssue)
			issue.PrintText("")
		}
	}
}

// PrintUsage prints store usage.
func (store *Jira2LevelsTreeStore) PrintUsage() {
	fmt.Println("Tasks store usage:")
	store.Level1Tasks.PrintUsage()
	fmt.Println("Sub Tasks store usage:")
	store.Level2Tasks.PrintUsage()
}

/*
Common
*/

func isLevel1Task(taskType string) bool {
	return taskType == "PMTask" || taskType == "Story"
}
