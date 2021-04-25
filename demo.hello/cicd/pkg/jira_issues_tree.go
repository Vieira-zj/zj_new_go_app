package pkg

import (
	"context"
	"fmt"
	"time"

	"demo.hello/utils"
)

// Tree interface for jira issues tree.
type Tree interface {
	GetIssueStore() *utils.Cache
	GetMRStore() *utils.Cache
	GetExpired() int64
	IsRunning() bool
	WaitDone()
	SubmitIssue(issueID string)
	GetEpics() map[string]struct{}
	GetStories() map[string]struct{}
}

// JiraIssuesTree collects jira issues and saves in store for search.
type JiraIssuesTree struct {
	ctx        context.Context
	parallel   int
	expired    int64
	running    bool
	epics      map[string]struct{}
	stories    map[string]struct{}
	issueQueue chan string
	mrQueue    chan string
	issueStore *utils.Cache
	mrStore    *utils.Cache
	jira       *JiraTool
	git        *GitlabTool
}

// NewJiraIssuesTree creates an instance of JiraIssuesTree.
func NewJiraIssuesTree(ctx context.Context, parallel int) *JiraIssuesTree {
	// 1.queueSize设置太小可能会导致阻塞 2.分片存储，mapSize不需要设置过大
	return &JiraIssuesTree{
		ctx:        ctx,
		parallel:   parallel,
		expired:    time.Now().Unix() + int64(expired),
		running:    false,
		epics:      make(map[string]struct{}, mapSize/2),
		stories:    make(map[string]struct{}, mapSize),
		issueQueue: make(chan string, queueSize),
		mrQueue:    make(chan string, queueSize),
		issueStore: utils.NewCache(parallel, mapSize),
		mrStore:    utils.NewCache(parallel, mapSize),
		jira:       NewJiraTool(),
		git:        NewGitlabTool(),
	}
}

// GetIssueStore .
func (tree *JiraIssuesTree) GetIssueStore() *utils.Cache {
	return tree.issueStore
}

// GetMRStore .
func (tree *JiraIssuesTree) GetMRStore() *utils.Cache {
	return tree.mrStore
}

// GetExpired .
func (tree *JiraIssuesTree) GetExpired() int64 {
	return tree.expired
}

// GetEpics .
func (tree *JiraIssuesTree) GetEpics() map[string]struct{} {
	return tree.epics
}

// GetStories .
func (tree *JiraIssuesTree) GetStories() map[string]struct{} {
	return tree.stories
}

// IsRunning .
func (tree *JiraIssuesTree) IsRunning() bool {
	return tree.running
}

// WaitDone .
func (tree *JiraIssuesTree) WaitDone() {
	for i := 0; i < 3; {
		if tree.queueSize() > 0 {
			i = 0
		} else {
			i++
		}
		time.Sleep(time.Second)
	}
	tree.running = false
}

// SubmitIssue puts a jira issue key in queue.
func (tree *JiraIssuesTree) SubmitIssue(issueID string) {
	tree.collect()
	tree.issueQueue <- issueID
}

func (tree *JiraIssuesTree) queueSize() int {
	return len(tree.issueQueue) + len(tree.mrQueue)
}

/*
Collect issues data.
*/

// collect fetches issues data and save in store.
func (tree *JiraIssuesTree) collect() {
	if tree.running {
		return
	}
	tree.running = true
	tree.collectIssues()
	tree.collectIssueMRs()
	// tree.addStorySuperIssues()
}

func (tree *JiraIssuesTree) collectIssues() {
	for i := 0; i < tree.parallel; i++ {
		go func() {
			var issueID string
			for {
				select {
				case issueID = <-tree.issueQueue:
					fmt.Println("Work on issue:", issueID)
				case <-tree.ctx.Done():
					fmt.Println("Issue worker exit.")
					return
				}

				issue := tree.collectOneIssue(issueID)
				if len(issue.Err) > 0 {
					continue
				}

				if issue.Type == issueTypeEpic {
					tree.epics[issueID] = struct{}{}
					// handle link story to epic
					for _, subIssueID := range issue.SubIssues {
						subIssue := tree.collectOneIssue(subIssueID)
						if len(subIssue.Err) > 0 {
							continue
						}
						subIssue.SuperIssues = append(issue.SuperIssues, issueID)
						tree.issueQueue <- subIssueID
					}
				} else if isStoryIssue(issue.Type) {
					tree.stories[issueID] = struct{}{}
					for _, subIssueID := range issue.SubIssues {
						if !tree.issueStore.IsExist(subIssueID) {
							tree.issueQueue <- subIssueID
						}
					}
				}
				for _, mrURL := range issue.MergeRequests {
					if !tree.mrStore.IsExist(mrURL) {
						tree.mrQueue <- mrURL
					}
				}
			}
		}()
	}
}

func (tree *JiraIssuesTree) collectOneIssue(issueID string) *JiraIssue {
	if value, err := tree.issueStore.Get(issueID); err == nil {
		return value.(*JiraIssue)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()
	issue, err := NewJiraIssueV2(ctx, tree.jira, issueID)
	if err != nil {
		content := fmt.Sprintf("New jira issue [%s] failed: %v\n", issueID, err)
		issue = &JiraIssue{Err: content}
	}
	tree.issueStore.PutIfEmpty(issueID, issue)
	return issue
}

func (tree *JiraIssuesTree) collectIssueMRs() {
	for i := 0; i < tree.parallel; i++ {
		go func() {
			var mrURL string
			for {
				select {
				case mrURL = <-tree.mrQueue:
					fmt.Println("Work on MR:", mrURL)
				case <-tree.ctx.Done():
					fmt.Println("MR worker exit.")
					return
				}

				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
				if mr, err := NewMergeRequest(ctx, tree.git, mrURL); err == nil {
					tree.mrStore.PutIfEmpty(mrURL, mr)
				} else {
					content := fmt.Sprintf("New merge request [%s] failed: %v\n", mrURL, err)
					mr := &MergeRequest{Err: content}
					if repo, err := getMRRepo(mrURL); err == nil {
						mr.Repo = repo
					}
					tree.mrStore.PutIfEmpty(mrURL, mr)
				}
				cancel()
			}
		}()
	}
}

// @deprecated: low performance
func (tree *JiraIssuesTree) addStorySuperIssues() {
	go func() {
		tree.WaitDone()
		for storyID := range tree.stories {
			value, err := tree.issueStore.Get(storyID)
			if err != nil {
				fmt.Printf("Get story [%s] failed: %v\n", storyID, err)
			}
			story := value.(*JiraIssue)
			for epicID := range tree.epics {
				value, err := tree.issueStore.Get(epicID)
				if err != nil {
					fmt.Printf("Get epic [%s] failed: %v\n", epicID, err)
				}
				isFound := false
				epic := value.(*JiraIssue)
				for _, subIssueID := range epic.SubIssues {
					if storyID == subIssueID {
						isFound = true
						story.SuperIssues = append(story.SuperIssues, epicID)
						break
					}
				}
				if isFound {
					break
				}
			}
		}
	}()
}

/*
Common
*/

func isStoryIssue(issueType string) bool {
	return issueType == issueTypePMTask || issueType == issueTypeStory
}

func removeDulpicatedItem(s []string) []string {
	m := make(map[string]struct{}, len(s))
	for _, item := range s {
		if _, ok := m[item]; ok {
			continue
		}
		m[item] = struct{}{}
	}

	retSlice := make([]string, 0, len(m))
	for key := range m {
		retSlice = append(retSlice, key)
	}
	return retSlice
}
