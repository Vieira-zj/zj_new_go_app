package pkg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"demo.hello/utils"
)

// Tree .
type Tree interface {
	GetIssueStore() *utils.Cache
	GetMRStore() *utils.Cache
	GetExpired() int64
	IsRunning() bool
	WaitDone()
	SubmitIssue(issueID string)
	getEpics() map[string]struct{}
	getStories() map[string]struct{}
}

// JiraIssuesTree handle jira issues and save in cache for search.
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

func (tree *JiraIssuesTree) getEpics() map[string]struct{} {
	return tree.epics
}

func (tree *JiraIssuesTree) getStories() map[string]struct{} {
	return tree.stories
}

func (tree *JiraIssuesTree) queueSize() int {
	return len(tree.issueQueue) + len(tree.mrQueue)
}

/*
Collect issues data.
*/

// collect fetches issue and merge request data.
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

				if issue.Type == "Epic" {
					tree.epics[issueID] = struct{}{}
					// handle link story to epic
					for _, subIssueID := range issue.SubIssues {
						subIssue := tree.collectOneIssue(subIssueID)
						if len(issue.Err) > 0 {
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
				} else if issue.Type == "Task" || issue.Type == "Bug" {
					for _, mrURL := range issue.MergeRequests {
						if !tree.mrStore.IsExist(mrURL) {
							tree.mrQueue <- mrURL
						}
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
				mr, err := NewMergeRequest(ctx, tree.git, mrURL)
				cancel()
				if err != nil {
					content := fmt.Sprintf("New merge request [%s] failed: %v\n", mrURL, err)
					tree.mrStore.PutIfEmpty(mrURL, &MergeRequest{Err: content})
					continue
				}
				tree.mrStore.PutIfEmpty(mrURL, mr)
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
Print text.
*/

// GetIssuesTreeText .
func GetIssuesTreeText(tree Tree) string {
	outLines := make([]string, 0, 100)
	outLines = append(outLines, "\n[Epic, Story and Tasks:]\n")
	for epicID := range tree.getEpics() {
		epic, epicText := GetIssueAndText(tree, epicID, "")
		outLines = append(outLines, epicText)
		if epic == nil {
			continue
		}
		for _, storyID := range epic.SubIssues {
			story, storyText := GetIssueAndText(tree, storyID, "\t")
			outLines = append(outLines, storyText)
			if story == nil {
				continue
			}
			for _, issueID := range story.SubIssues {
				_, issueText := GetIssueAndText(tree, issueID, "\t\t")
				outLines = append(outLines, issueText)
			}
		}
		outLines = append(outLines, "\n")
	}

	outLines = append(outLines, "\n[Single Stories:]\n")
	for storyID := range tree.getStories() {
		story, storyText := GetIssueAndText(tree, storyID, "")
		if len(story.SuperIssues) > 0 {
			continue
		}
		outLines = append(outLines, storyText)
		if story == nil {
			continue
		}
		for _, issueID := range story.SubIssues {
			_, issueText := GetIssueAndText(tree, issueID, "\t")
			outLines = append(outLines, issueText)
		}
		outLines = append(outLines, "\n")
	}

	outLines = append(outLines, "\n[Single Issues:]\n")
	for _, value := range tree.GetIssueStore().GetItems() {
		issue := value.(*JiraIssue)
		found := false
		for _, supIssueID := range issue.SuperIssues {
			if _, err := tree.GetIssueStore().Get(supIssueID); err == nil {
				found = true
				break
			}
		}
		if found {
			continue
		}

		if issue.Type == "Task" || issue.Type == "Bug" {
			if len(issue.Err) == 0 {
				outLines = append(outLines, issue.ToText())
				outLines = append(outLines, getIssueMRsText(tree, issue, "\t"))
			} else {
				outLines = append(outLines, issue.Err)
			}
			outLines = append(outLines, "\n")
		}
	}
	return strings.Join(outLines, "")
}

// GetIssueAndText .
func GetIssueAndText(tree Tree, issueID string, prefix string) (*JiraIssue, string) {
	value, err := tree.GetIssueStore().Get(issueID)
	if err != nil {
		return nil, fmt.Sprintf("%sGet issue [%s] failed: %v\n", prefix, issueID, err)
	}
	issue := value.(*JiraIssue)
	if len(issue.Err) > 0 {
		return nil, prefix + issue.Err
	}

	retlines := make([]string, 10)
	retlines = append(retlines, prefix+issue.ToText())
	retlines = append(retlines, getIssueMRsText(tree, issue, prefix+"\t"))
	return issue, strings.Join(retlines, "")
}

func getIssueMRsText(tree Tree, issue *JiraIssue, prefix string) string {
	if isStoryIssue(issue.Type) {
		return ""
	}

	outLines := make([]string, 10)
	usedMR := make(map[string]struct{})
	for _, mrURL := range issue.MergeRequests {
		if _, ok := usedMR[mrURL]; ok {
			continue
		}
		usedMR[mrURL] = struct{}{}

		value, err := tree.GetMRStore().Get(mrURL)
		if err != nil {
			line := fmt.Sprintf("%sGet mr [%s] failed: %v\n", prefix, mrURL, err)
			outLines = append(outLines, line)
			continue
		}
		mr := value.(*MergeRequest)
		if len(mr.Err) > 0 {
			outLines = append(outLines, prefix+mr.Err)
			continue
		}

		if mr.TargetBR == "master" {
			outLines = append(outLines, prefix+mr.ToText())
		}
	}
	return strings.Join(outLines, "")
}

// GetIssuesTreeUsageText returns tree store usage.
func GetIssuesTreeUsageText(tree Tree) string {
	outLines := make([]string, 0, 5)
	outLines = append(outLines, "Tree issue store usage:\n")
	outLines = append(outLines, tree.GetIssueStore().UsageToText())
	outLines = append(outLines, "Tree MR store usage:\n")
	outLines = append(outLines, tree.GetMRStore().UsageToText())
	return strings.Join(outLines, "")
}

// GetIssuesTreeSummary .
func GetIssuesTreeSummary(tree Tree) string {
	outLines := make([]string, 0, 5)
	outLines = append(outLines, fmt.Sprintf("Total Epics: %d\n", len(tree.getEpics())))
	outLines = append(outLines, fmt.Sprintf("Total Stories (PMTask): %d\n", len(tree.getStories())))
	totalTasks := tree.GetIssueStore().Size() - len(tree.getEpics()) - len(tree.getStories())
	outLines = append(outLines, fmt.Sprintf("Total Tasks (include Bugs): %d\n", totalTasks))
	return strings.Join(outLines, "")
}

/*
Common
*/

func isStoryIssue(issueType string) bool {
	return issueType == "PMTask" || issueType == "Story"
}
