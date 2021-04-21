package pkg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"demo.hello/utils"
)

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
	const (
		queueSize = 30
		mapSize   = 20
	)
	return &JiraIssuesTree{
		ctx:        ctx,
		parallel:   parallel,
		expired:    time.Now().Unix() + int64(expired),
		running:    false,
		epics:      make(map[string]struct{}, mapSize/2),
		stories:    make(map[string]struct{}, mapSize),
		issueQueue: make(chan string, queueSize),
		mrQueue:    make(chan string, queueSize),
		issueStore: utils.NewCache((parallel * 2), mapSize),
		mrStore:    utils.NewCache(parallel, (mapSize / 2)),
		jira:       NewJiraTool(),
		git:        NewGitlabTool(),
	}
}

// GetStore returns internal store.
func (tree *JiraIssuesTree) GetStore() *utils.Cache {
	return tree.issueStore
}

// GetExpired .
func (tree *JiraIssuesTree) GetExpired() int64 {
	return tree.expired
}

// IsRunning .
func (tree *JiraIssuesTree) IsRunning() bool {
	return tree.running
}

// QueueSize returns total issue keys to be handle in queue.
func (tree *JiraIssuesTree) QueueSize() int {
	return len(tree.issueQueue) + len(tree.mrQueue)
}

/*
Collect issues data.
*/

// SubmitIssue puts a jira issue key in queue.
func (tree *JiraIssuesTree) SubmitIssue(issueID string) error {
	tree.collect()
	tree.issueQueue <- issueID
	return nil
}

// collect fetches issue and merge request data.
func (tree *JiraIssuesTree) collect() {
	if tree.running {
		return
	}
	tree.running = true
	tree.collectIssues()
	tree.collectIssueMRs()
	// tree.addStorySuperIssues()
	tree.waitDone()
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
					fmt.Println("Work on mr:", mrURL)
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
		time.Sleep(time.Second)
		for tree.QueueSize() > 0 {
			time.Sleep(time.Second)
		}

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

func (tree *JiraIssuesTree) waitDone() {
	go func() {
		for i := 0; i < 3; {
			if tree.QueueSize() > 0 {
				i = 0
			} else {
				i++
			}
			time.Sleep(time.Second)
		}
		tree.running = false
	}()
}

/*
Print text.
*/

// ToText .
func (tree *JiraIssuesTree) ToText() string {
	outLines := make([]string, 0, 100)
	outLines = append(outLines, "\n[Epic, Story and Tasks:]\n")
	for epicID := range tree.epics {
		epic, epicText := tree.GetIssueAndText(epicID, "")
		outLines = append(outLines, epicText)
		if epic == nil {
			continue
		}
		for _, storyID := range epic.SubIssues {
			story, storyText := tree.GetIssueAndText(storyID, "\t")
			outLines = append(outLines, storyText)
			if story == nil {
				continue
			}
			for _, issueID := range story.SubIssues {
				_, issueText := tree.GetIssueAndText(issueID, "\t\t")
				outLines = append(outLines, issueText)
			}
		}
		outLines = append(outLines, "\n")
	}

	outLines = append(outLines, "\n[Single Stories:]\n")
	for storyID := range tree.stories {
		story, storyText := tree.GetIssueAndText(storyID, "")
		if len(story.SuperIssues) > 0 {
			continue
		}
		outLines = append(outLines, storyText)
		if story == nil {
			continue
		}
		for _, issueID := range story.SubIssues {
			_, issueText := tree.GetIssueAndText(issueID, "\t")
			outLines = append(outLines, issueText)
		}
		outLines = append(outLines, "\n")
	}

	outLines = append(outLines, "\n[Single Issues:]\n")
	for _, value := range tree.issueStore.GetItems() {
		issue := value.(*JiraIssue)
		found := false
		for _, supIssueID := range issue.SuperIssues {
			if _, err := tree.issueStore.Get(supIssueID); err == nil {
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
				outLines = append(outLines, tree.issueMRsToText(issue, "\t"))
			} else {
				outLines = append(outLines, issue.Err)
			}
			outLines = append(outLines, "\n")
		}
	}
	return strings.Join(outLines, "")
}

// GetIssueAndText .
func (tree *JiraIssuesTree) GetIssueAndText(issueID string, prefix string) (*JiraIssue, string) {
	value, err := tree.issueStore.Get(issueID)
	if err != nil {
		return nil, fmt.Sprintf("%sGet issue [%s] failed: %v\n", prefix, issueID, err)
	}
	issue := value.(*JiraIssue)
	if len(issue.Err) > 0 {
		return nil, prefix + issue.Err
	}

	retlines := make([]string, 10)
	retlines = append(retlines, prefix+issue.ToText())
	retlines = append(retlines, tree.issueMRsToText(issue, prefix+"\t"))
	return issue, strings.Join(retlines, "")
}

func (tree *JiraIssuesTree) issueMRsToText(issue *JiraIssue, prefix string) string {
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

		value, err := tree.mrStore.Get(mrURL)
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

// PrintUsage prints tree store usage.
func (tree *JiraIssuesTree) PrintUsage() {
	fmt.Println(tree.UsageToText())
}

// UsageToText returns tree store usage.
func (tree *JiraIssuesTree) UsageToText() string {
	outLines := make([]string, 0, 5)
	outLines = append(outLines, "Tree issue store usage:\n")
	outLines = append(outLines, tree.issueStore.UsageToText())
	outLines = append(outLines, "Tree mr store usage:\n")
	outLines = append(outLines, tree.mrStore.UsageToText())
	return strings.Join(outLines, "")
}

/*
Common
*/

func isStoryIssue(issueType string) bool {
	return issueType == "PMTask" || issueType == "Story"
}
