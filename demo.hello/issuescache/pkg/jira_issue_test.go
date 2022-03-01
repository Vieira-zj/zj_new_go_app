package pkg

import (
	"context"
	"fmt"
	"testing"
)

var issueKey = "SPPAY-196"

func TestNewJiraIssue(t *testing.T) {
	jira := NewJiraTool()
	issue, err := NewJiraIssue(context.TODO(), jira, issueKey)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(issue.ToText())
}

func TestNewJiraIssueV2(t *testing.T) {
	jira := NewJiraTool()
	issue, err := NewJiraIssueV2(context.TODO(), jira, issueKey)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(issue.ToText())
}
