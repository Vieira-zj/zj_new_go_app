package pkg

import (
	"context"
	"fmt"
	"time"
)

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

	tree := NewJiraIssuesTreeV2(c.parallel)
	for _, key := range keys {
		tree.SubmitIssue(key)
	}

	tree.WaitDone()
	fmt.Println(GetIssuesTreeText(tree))
	fmt.Println(GetIssuesTreeUsageText(tree))
	return nil
}
