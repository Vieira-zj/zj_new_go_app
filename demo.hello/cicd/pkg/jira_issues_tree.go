package pkg

import (
	"context"
	"fmt"
	"strings"
	"time"

	"demo.hello/utils"
)

/*
Worker
*/

// JiraIssuesTree handle jira issues and save in cache for search.
type JiraIssuesTree struct {
	Key        string
	parallel   int
	ctx        context.Context
	roots      []string
	issueQueue chan string
	issueStore *utils.Cache
	mrQueue    chan string
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
		roots:      make([]string, 0, parallel*mapSize),
		issueQueue: make(chan string, queueSize),
		issueStore: utils.NewCache((parallel * 2), mapSize),
		mrQueue:    make(chan string, queueSize),
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
					fmt.Printf("New jira issue [%s] failed: %v\n", issueID, err)
					continue
				}

				tree.issueStore.PutIfEmpty(issueID, issue)
				if isStoryIssue(issue.Type) {
					tree.roots = append(tree.roots, issueID)
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

				ctx, cancel := context.WithTimeout(context.Background(), time.Duration(8)*time.Second)
				mr, err := NewMergeRequest(ctx, tree.git, mrURL)
				cancel()
				if err != nil {
					fmt.Printf("New merge request [%s] failed: %v\n", mrURL, err)
					continue
				}

				if mr.TargetBR == "master" {
					tree.mrStore.PutIfEmpty(mrURL, mr)
				}
			}
		}()
	}
}

// PrintText prints 2level tree (PMTask->DevTask, Story->Task) as text.
func (tree *JiraIssuesTree) PrintText() {
	fmt.Println(tree.ToText())
}

// ToText .
func (tree *JiraIssuesTree) ToText() string {
	outLines := make([]string, 0, 100)
	var usedIssues map[string]struct{}
	if len(tree.roots) > 0 {
		usedIssues = make(map[string]struct{}, 10)
		outLines = append(outLines, "\n[Issues and Sub Issues:]\n")
		for _, issueID := range tree.roots {
			value, err := tree.issueStore.Get(issueID)
			if err != nil {
				outLines = append(outLines, fmt.Sprintf("Get issue [%s] failed: %v\n", issueID, err))
				continue
			}
			issue := value.(*JiraIssue)
			outLines = append(outLines, issue.ToText())
			for _, subIssueID := range issue.SubIssues {
				usedIssues[subIssueID] = struct{}{}
				if subValue, err := tree.issueStore.Get(subIssueID); err != nil {
					outLines = append(outLines, fmt.Sprintf("Get sub issue [%s] failed: %v\n", subIssueID, err))
				} else {
					subIssue := subValue.(*JiraIssue)
					outLines = append(outLines, "\t"+subIssue.ToText())
					outLines = append(outLines, tree.issueMRsToText(subIssue, "\t\t"))
				}
			}
			outLines = append(outLines, "\n")
		}
	}

	outLines = append(outLines, "\n[Single Issues:]\n")
	for _, value := range tree.issueStore.GetItems() {
		issue := value.(*JiraIssue)
		if !isStoryIssue(issue.Type) {
			if _, ok := usedIssues[issue.Key]; !ok {
				outLines = append(outLines, issue.ToText())
				outLines = append(outLines, tree.issueMRsToText(issue, "\t"))
				outLines = append(outLines, "\n")
			}
		}
	}
	return strings.Join(outLines, "")
}

// IssueToText .
func (tree *JiraIssuesTree) IssueToText(issue *JiraIssue, prefix string) string {
	outLines := make([]string, 0, 10)
	outLines = append(outLines, issue.ToText())
	for _, subIssueID := range issue.SubIssues {
		if subValue, err := tree.issueStore.Get(subIssueID); err != nil {
			outLines = append(outLines, fmt.Sprintf("Get sub issue [%s] failed: %v\n", subIssueID, err))
		} else {
			subIssue := subValue.(*JiraIssue)
			outLines = append(outLines, "\t"+subIssue.ToText())
			outLines = append(outLines, tree.issueMRsToText(subIssue, "\t\t"))
		}
	}
	return strings.Join(outLines, "")
}

func (tree *JiraIssuesTree) issueMRsToText(issue *JiraIssue, prefix string) string {
	outLines := make([]string, 10)
	for _, mrURL := range issue.MergeRequests {
		value, err := tree.mrStore.Get(mrURL)
		if err != nil {
			continue
		}
		outLines = append(outLines, prefix+value.(*MergeRequest).ToText())
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
