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

// JiraIssuesTree handle jira issues and save in cache for search.
type JiraIssuesTree struct {
	Key      string
	parallel int
	ctx      context.Context
	queue    chan string
	jira     *JiraTool
	roots    []string
	store    *utils.Cache
}

// NewJiraIssuesTree creates an instance of JiraIssuesTree.
func NewJiraIssuesTree(ctx context.Context, key string, parallel int) *JiraIssuesTree {
	// 1.queueSize设置太小可能会导致阻塞 2.分片存储，mapSize不需要设置过大
	const (
		queueSize = 30
		mapSize   = 20
	)
	return &JiraIssuesTree{
		Key:      key,
		parallel: parallel,
		ctx:      ctx,
		queue:    make(chan string, queueSize),
		jira:     NewJiraTool(),
		roots:    make([]string, 0, parallel*mapSize),
		store:    utils.NewCache((parallel * 2), mapSize),
	}
}

// QueueSize returns total issue keys to be handle in queue.
func (tree *JiraIssuesTree) QueueSize() int {
	return len(tree.queue)
}

// GetStore returns internal store.
func (tree *JiraIssuesTree) GetStore() *utils.Cache {
	return tree.store
}

// Submit puts a jira issue key in queue.
func (tree *JiraIssuesTree) Submit(issueID string) {
	tree.queue <- issueID
}

// CollectIssues handles jira issues and save in store.
func (tree *JiraIssuesTree) CollectIssues() {
	for i := 0; i < tree.parallel; i++ {
		go func() {
			var issueID string
			for {
				select {
				case issueID = <-tree.queue:
					fmt.Println("Work on issue:", issueID)
				case <-tree.ctx.Done():
					fmt.Println("Worker exit.")
					return
				}

				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
				issue, err := NewJiraIssueV2(ctx, tree.jira, issueID)
				cancel()
				if err != nil {
					fmt.Println("Create jira issue failed:", err)
					continue
				}

				tree.store.Put(issueID, issue)
				if isStoryIssue(issue.Type) {
					tree.roots = append(tree.roots, issueID)
					for _, subIssueID := range issue.SubIssues {
						if !tree.store.IsExist(subIssueID) {
							tree.queue <- subIssueID
						}
					}
				}
			}
		}()
	}
}

// PrintText prints 2level tree (PMTask->DevTask, Story->Task) as text.
func (tree *JiraIssuesTree) PrintText() {
	var usedIssues map[string]struct{}
	if len(tree.roots) > 0 {
		usedIssues = make(map[string]struct{}, 10)
		fmt.Println("\n[Issues and Sub Issues:]")
		for _, issueID := range tree.roots {
			item, err := tree.store.Get(issueID)
			if err != nil {
				fmt.Printf("Get issue [%s] failed: %v\n", issueID, err)
				continue
			}
			issue := item.(*JiraIssue)
			issue.PrintText("")
			for _, subIssueID := range issue.SubIssues {
				usedIssues[subIssueID] = struct{}{}
				if subIssue, err := tree.store.Get(subIssueID); err != nil {
					fmt.Printf("Get sub issue [%s] failed: %v\n", subIssueID, err)
				} else {
					subIssue.(*JiraIssue).PrintText("\t")
				}
			}
			fmt.Println()
		}
	}

	fmt.Println("\n[Single Issues:]")
	for _, item := range tree.store.GetItems() {
		issue := item.(*JiraIssue)
		if !isStoryIssue(issue.Type) {
			if _, ok := usedIssues[issue.Key]; !ok {
				issue.PrintText("")
			}
		}
	}
}

// PrintUsage prints tree store usage.
func (tree *JiraIssuesTree) PrintUsage() {
	fmt.Println("Tree store usage:")
	tree.store.PrintUsage()
}

/*
Common
*/

func isStoryIssue(issueType string) bool {
	return issueType == "PMTask" || issueType == "Story"
}
