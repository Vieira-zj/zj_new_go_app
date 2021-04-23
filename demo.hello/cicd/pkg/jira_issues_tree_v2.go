package pkg

import (
	"context"
	"fmt"
	"sync"
	"time"

	"demo.hello/utils"
)

// JiraIssuesTreeV2 .
type JiraIssuesTreeV2 struct {
	ctx        context.Context
	wg         sync.WaitGroup
	expired    int64
	running    bool
	epics      map[string]struct{}
	stories    map[string]struct{}
	semaphore  chan struct{}
	issueStore *utils.Cache
	mrStore    *utils.Cache
	jira       *JiraTool
	git        *GitlabTool
}

// NewJiraIssuesTreeV2 .
func NewJiraIssuesTreeV2(ctx context.Context, parallel int) *JiraIssuesTreeV2 {
	return &JiraIssuesTreeV2{
		ctx:        ctx,
		wg:         sync.WaitGroup{},
		expired:    time.Now().Unix() + int64(expired),
		running:    false,
		epics:      make(map[string]struct{}, mapSize/2),
		stories:    make(map[string]struct{}, mapSize),
		semaphore:  make(chan struct{}, parallel),
		issueStore: utils.NewCache(parallel, mapSize),
		mrStore:    utils.NewCache(parallel, mapSize),
		jira:       NewJiraTool(),
		git:        NewGitlabTool(),
	}
}

// GetIssueStore .
func (tree *JiraIssuesTreeV2) GetIssueStore() *utils.Cache {
	return tree.issueStore
}

// GetMRStore .
func (tree *JiraIssuesTreeV2) GetMRStore() *utils.Cache {
	return tree.mrStore
}

// GetExpired .
func (tree *JiraIssuesTreeV2) GetExpired() int64 {
	return tree.expired
}

// IsRunning .
func (tree *JiraIssuesTreeV2) IsRunning() bool {
	return tree.running
}

// WaitDone .
func (tree *JiraIssuesTreeV2) WaitDone() {
	if tree.running {
		tree.wg.Wait()
		tree.running = false
	}
}

// SubmitIssue .
func (tree *JiraIssuesTreeV2) SubmitIssue(issueID string) {
	tree.running = true
	tree.collectIssues(issueID)
}

func (tree *JiraIssuesTreeV2) getEpics() map[string]struct{} {
	return tree.epics
}

func (tree *JiraIssuesTreeV2) getStories() map[string]struct{} {
	return tree.stories
}

/*
Collect data.
*/

func (tree *JiraIssuesTreeV2) collectIssues(issueID string) {
	tree.wg.Add(1)
	go func() {
		fmt.Println("Work on issue:", issueID)
		tree.semaphore <- struct{}{}
		defer func() {
			tree.wg.Done()
			<-tree.semaphore
		}()

		issue := tree.collectOneIssue(issueID)
		if issue.Type == "Epic" {
			tree.epics[issueID] = struct{}{}
			for _, subIssueID := range issue.SubIssues {
				subIssue := tree.collectOneIssue(subIssueID)
				subIssue.SuperIssues = append(subIssue.SuperIssues, issueID)
				tree.collectIssues(subIssueID)
			}
		} else if isStoryIssue(issue.Type) {
			tree.stories[issueID] = struct{}{}
			for _, subIssueID := range issue.SubIssues {
				tree.collectIssues(subIssueID)
			}
		} else if issue.Type == "Task" || issue.Type == "Bug" {
			for _, mrURL := range issue.MergeRequests {
				tree.collectIssueMRs(mrURL)
			}
		}
	}()
}

func (tree *JiraIssuesTreeV2) collectOneIssue(issueID string) *JiraIssue {
	if value, err := tree.issueStore.Get(issueID); err == nil {
		return value.(*JiraIssue)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	issue, err := NewJiraIssueV2(ctx, tree.jira, issueID)
	cancel()
	if err != nil {
		content := fmt.Sprintf("New jira issue [%s] failed: %v\n", issueID, err)
		issue = &JiraIssue{Err: content}
	}

	tree.issueStore.PutIfEmpty(issueID, issue)
	return issue
}

func (tree *JiraIssuesTreeV2) collectIssueMRs(mrURL string) {
	tree.wg.Add(1)
	go func() {
		fmt.Println("Work on MR:", mrURL)
		tree.semaphore <- struct{}{}
		defer func() {
			tree.wg.Done()
			<-tree.semaphore
		}()

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		mr, err := NewMergeRequest(ctx, tree.git, mrURL)
		cancel()
		if err != nil {
			content := fmt.Sprintf("New merge request [%s] failed: %v\n", mrURL, err)
			tree.mrStore.PutIfEmpty(mrURL, &MergeRequest{Err: content})
		}
		tree.mrStore.PutIfEmpty(mrURL, mr)
	}()
}
