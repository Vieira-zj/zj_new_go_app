package pkg

import (
	"context"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"
)

type byIssueKey []string

func (keys byIssueKey) Len() int {
	return len(keys)
}

func (keys byIssueKey) Swap(i, j int) {
	keys[i], keys[j] = keys[j], keys[i]
}

func (keys byIssueKey) Less(i, j int) bool {
	iItems := strings.Split(keys[i], "-")
	jItems := strings.Split(keys[j], "-")
	if iItems[0] != jItems[0] {
		if iItems[0][0] == 'A' {
			return true
		}
		return false
	}

	iID, err := strconv.Atoi(iItems[1])
	if err != nil {
		fmt.Println(err) // not go here
	}
	jID, err := strconv.Atoi(jItems[1])
	if err != nil {
		fmt.Println(err)
	}
	return iID < jID
}

// Cmd .
type Cmd struct {
	jira     *JiraTool
	parallel int
}

// NewCmd .
func NewCmd(parallel int) *Cmd {
	return &Cmd{
		jira:     NewJiraTool(),
		parallel: parallel,
	}
}

// PrintFixVersionTree .
func (c *Cmd) PrintFixVersionTree(version string) error {
	jql := fmt.Sprintf("fixVersion = %s", version)
	return c.PrintJiraIssuesTree(jql)
}

// PrintReleaseCycleTree .
func (c *Cmd) PrintReleaseCycleTree(rc string) error {
	jql := fmt.Sprintf(`"Release Cycle" = "%s"`, rc)
	return c.PrintJiraIssuesTree(jql)
}

// PrintJiraIssuesTree .
func (c *Cmd) PrintJiraIssuesTree(jql string) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(8)*time.Second)
	defer cancel()

	keys, err := c.jira.SearchIssues(ctx, jql)
	if err != nil {
		return err
	}
	sort.Sort(byIssueKey(keys))
	fmt.Println("Jira issues to be handle:", strings.Join(keys, ","))

	tree := NewJiraIssuesTreeV2(c.parallel)
	for _, key := range keys {
		tree.SubmitIssue(key)
	}

	tree.WaitDone()
	fmt.Println(GetIssuesTreeText(tree))
	fmt.Println(GetIssuesTreeUsageText(tree))
	return nil
}
