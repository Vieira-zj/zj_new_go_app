package pkg

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestGetDeployReposText(t *testing.T) {
	jira := NewJiraTool()
	jql := `"Release Cycle" = "2021.04.v4 - Payment"`
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(8)*time.Second)
	defer cancel()
	keys, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		t.Fatal(err)
	}

	tree := NewJiraIssuesTreeV2(10)
	for _, key := range keys {
		tree.SubmitIssue(key)
	}
	tree.WaitDone()

	fmt.Println(GetIssuesTreeText(tree))

	fmt.Println(GetDeployReposText(tree))
	fmt.Println(GetShortDeployReposText(tree))

	fmt.Println(GetIssuesTreeUsageText(tree))
	fmt.Println(GetIssuesTreeSummaryText(tree))
}
