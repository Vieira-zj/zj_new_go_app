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
	roots      map[string]struct{}
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
		parallel:   parallel,
		ctx:        ctx,
		roots:      make(map[string]struct{}, parallel*mapSize),
		issueQueue: make(chan string, queueSize),
		mrQueue:    make(chan string, queueSize),
		issueStore: utils.NewCache((parallel * 2), mapSize),
		mrStore:    utils.NewCache(parallel, (mapSize / 2)),
		jira:       NewJiraTool(),
		git:        NewGitlabTool(),
	}
}

// QueueSize returns total issue keys to be handle in queue.
func (tree *JiraIssuesTree) QueueSize() int {
	return len(tree.issueQueue) + len(tree.mrQueue)
}

// GetStore returns internal store.
func (tree *JiraIssuesTree) GetStore() *utils.Cache {
	return tree.issueStore
}

// SubmitIssue puts a jira issue key in queue.
func (tree *JiraIssuesTree) SubmitIssue(issueID string) {
	tree.issueQueue <- issueID
}

// Collect .
func (tree *JiraIssuesTree) Collect() {
	tree.collectIssues()
	tree.collectMRs()
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

				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
				issue, err := NewJiraIssueV2(ctx, tree.jira, issueID)
				cancel()
				if err != nil {
					content := fmt.Sprintf("New jira issue [%s] failed: %v\n", issueID, err)
					tree.issueStore.PutIfEmpty(issueID, &JiraIssue{Err: content})
					continue
				}

				tree.issueStore.PutIfEmpty(issueID, issue)
				if isStoryIssue(issue.Type) {
					tree.roots[issueID] = struct{}{}
					for _, subIssueID := range issue.SubIssues {
						if !tree.issueStore.IsExist(subIssueID) {
							tree.issueQueue <- subIssueID
						}
					}
				} else {
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

func (tree *JiraIssuesTree) collectMRs() {
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

// ToText .
func (tree *JiraIssuesTree) ToText() string {
	outLines := make([]string, 0, 100)
	var usedIssues map[string]struct{}
	if len(tree.roots) > 0 {
		usedIssues = make(map[string]struct{}, 10)
		outLines = append(outLines, "\n[Issues and Sub Issues:]\n")
		for issueID := range tree.roots {
			issue, issueText := tree.GetIssueAndText(issueID, "")
			outLines = append(outLines, issueText)
			if issue == nil {
				continue
			}

			for _, subIssueID := range issue.SubIssues {
				usedIssues[subIssueID] = struct{}{}
				_, subIssueText := tree.GetIssueAndText(subIssueID, "\t")
				outLines = append(outLines, subIssueText)
			}
			outLines = append(outLines, "\n")
		}
	}

	outLines = append(outLines, "\n[Single Issues:]\n")
	for key, value := range tree.issueStore.GetItems() {
		if _, ok := tree.roots[key]; ok {
			continue
		}
		if _, ok := usedIssues[key]; ok {
			continue
		}

		issue := value.(*JiraIssue)
		if len(issue.Err) == 0 {
			outLines = append(outLines, issue.ToText())
			outLines = append(outLines, tree.issueMRsToText(issue, "\t"))
		} else {
			outLines = append(outLines, issue.Err)
		}
		outLines = append(outLines, "\n")
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
