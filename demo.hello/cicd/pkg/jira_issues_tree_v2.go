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
func NewJiraIssuesTreeV2(parallel int) *JiraIssuesTreeV2 {
	return &JiraIssuesTreeV2{
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

// GetEpics .
func (tree *JiraIssuesTreeV2) GetEpics() map[string]struct{} {
	return tree.epics
}

// GetStories .
func (tree *JiraIssuesTreeV2) GetStories() map[string]struct{} {
	return tree.stories
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

/*
Collect issue data.
*/

// collectIssues fetches issues data and save in store.
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
		if len(issue.Err) > 0 {
			return
		}

		if issue.Type == issueTypeEpic {
			tree.epics[issueID] = struct{}{}
			for _, subIssueID := range issue.SubIssues {
				subIssue := tree.collectOneIssue(subIssueID)
				if len(subIssue.Err) > 0 {
					continue
				}
				subIssue.SuperIssues = append(subIssue.SuperIssues, issueID)
				tree.collectIssues(subIssueID)
			}
		} else if isStoryIssue(issue.Type) {
			tree.stories[issueID] = struct{}{}
			for _, subIssueID := range issue.SubIssues {
				tree.collectIssues(subIssueID)
			}
		}
		for _, mrURL := range issue.MergeRequests {
			tree.collectIssueMRs(issueID, mrURL)
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
		content := fmt.Sprintf("New jira issue failed: %v\n", err)
		issue = &JiraIssue{Key: issueID, Err: content}
	}
	tree.issueStore.PutIfEmpty(issueID, issue)
	return issue
}

func (tree *JiraIssuesTreeV2) collectIssueMRs(issueID, mrURL string) {
	tree.wg.Add(1)
	go func() {
		fmt.Printf("Work on MR (linked issue %s): %s\n", issueID, mrURL)
		tree.semaphore <- struct{}{}
		defer func() {
			tree.wg.Done()
			<-tree.semaphore
		}()

		ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
		defer cancel()
		if mr, err := NewMergeRequest(ctx, tree.git, mrURL); err == nil {
			mr.LinkedIssue = issueID
			tree.mrStore.PutIfEmpty(mrURL, mr)
		} else {
			content := fmt.Sprintf("New merge request failed: %v\n", err)
			mr := &MergeRequest{WebURL: mrURL, Err: content}
			if repo, err := getMRRepo(mrURL); err == nil {
				mr.Repo = repo
			}
			tree.mrStore.PutIfEmpty(mrURL, mr)
		}
	}()
}
