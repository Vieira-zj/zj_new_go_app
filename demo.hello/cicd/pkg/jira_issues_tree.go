package pkg

import (
	"context"
	"fmt"
	"strings"
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
					tree.mrStore.PutIfEmpty(mrURL, &MergeRequest{Err: content})
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
Print tree text.
*/

// GetIssuesTreeText .
func GetIssuesTreeText(tree Tree) string {
	outLines := make([]string, 0, 100)
	errLines := make([]string, 0, 10)
	errLines = append(errLines, "\n[Failed Tasks:]\n")

	outLines = append(outLines, "\n[Epic, Story and Tasks:]\n")
	for epicID := range tree.GetEpics() {
		epic, epicText := GetIssueAndMRsText(tree, epicID, "")
		if epic == nil {
			errLines = append(errLines, epicText)
			continue
		}
		outLines = append(outLines, epicText)
		for _, storyID := range epic.SubIssues {
			story, storyText := GetIssueAndMRsText(tree, storyID, "\t")
			if story == nil {
				errLines = append(errLines, storyText)
				continue
			}
			outLines = append(outLines, storyText)
			for _, issueID := range story.SubIssues {
				if issue, issueText := GetIssueAndMRsText(tree, issueID, "\t\t"); issue == nil {
					errLines = append(errLines, issueText)
				} else {
					outLines = append(outLines, issueText)
				}
			}
		}
		outLines = append(outLines, "\n")
	}

	outLines = append(outLines, "\n[Single Stories:]\n")
	for storyID := range tree.GetStories() {
		story, storyText := GetIssueAndMRsText(tree, storyID, "")
		if story == nil {
			errLines = append(errLines, storyText)
			continue
		}
		if isDulplicatedIssue(tree, story) {
			continue
		}
		outLines = append(outLines, storyText)
		for _, issueID := range story.SubIssues {
			if issue, issueText := GetIssueAndMRsText(tree, issueID, "\t"); issue == nil {
				errLines = append(errLines, issueText)
			} else {
				outLines = append(outLines, issueText)
			}
		}
		outLines = append(outLines, "\n")
	}

	outLines = append(outLines, "\n[Single Tasks (Bugs):]\n")
	for issueID := range tree.GetIssueStore().GetItems() {
		if issue, issueText := GetIssueAndMRsText(tree, issueID, ""); issue != nil {
			if (issue.Type == issueTypeTask || issue.Type == issueTypeBug) && !isDulplicatedIssue(tree, issue) {
				outLines = append(outLines, issueText)
				outLines = append(outLines, "\n")
			}
		} else {
			errLines = append(errLines, issueText)
		}
	}

	outLines = append(outLines, removeDulpicatedItem(errLines)...)
	return strings.Join(outLines, "")
}

// GetIssueAndMRsText .
func GetIssueAndMRsText(tree Tree, issueID string, prefix string) (*JiraIssue, string) {
	value, err := tree.GetIssueStore().Get(issueID)
	if err != nil {
		return nil, fmt.Sprintf("%sGet issue [%s] from store failed: %v\n", prefix, issueID, err)
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

// GetIssuesTreeUsageText returns tree store usage.
func GetIssuesTreeUsageText(tree Tree) string {
	outLines := make([]string, 0, 5)
	outLines = append(outLines, "Tree issue store usage:\n")
	outLines = append(outLines, tree.GetIssueStore().UsageToText())
	outLines = append(outLines, "Tree MR store usage:\n")
	outLines = append(outLines, tree.GetMRStore().UsageToText())
	return strings.Join(outLines, "")
}

// GetIssuesTreeSummaryText .
func GetIssuesTreeSummaryText(tree Tree) string {
	outLines := make([]string, 0, 5)
	outLines = append(outLines, fmt.Sprintf("Total Epics: %d\n", len(tree.GetEpics())))
	outLines = append(outLines, fmt.Sprintf("Total Stories (PMTask): %d\n", len(tree.GetStories())))
	totalTasks := tree.GetIssueStore().Size() - len(tree.GetEpics()) - len(tree.GetStories())
	outLines = append(outLines, fmt.Sprintf("Total Tasks (include Bugs): %d\n", totalTasks))
	return strings.Join(outLines, "")
}

func getIssueMRsText(tree Tree, issue *JiraIssue, prefix string) string {
	outLines := make([]string, 5)
	dulplicatedMR := make(map[string]struct{}, 5)
	for _, mrURL := range issue.MergeRequests {
		if _, ok := dulplicatedMR[mrURL]; ok {
			continue
		}
		dulplicatedMR[mrURL] = struct{}{}

		value, err := tree.GetMRStore().Get(mrURL)
		if err != nil {
			line := fmt.Sprintf("%sGet MR [%s] from store failed: %v\n", prefix, mrURL, err)
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

func isDulplicatedIssue(tree Tree, issue *JiraIssue) bool {
	for _, supIssueID := range issue.SuperIssues {
		if _, err := tree.GetIssueStore().Get(supIssueID); err == nil {
			return true
		}
	}
	return false
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
