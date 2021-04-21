package pkg

import (
	"context"
	"testing"
)

var issueKey = "SPPAY-196"

func TestNewJiraIssue(t *testing.T) {
	issue, err := NewJiraIssue(context.TODO(), jira, issueKey)
	if err != nil {
		t.Fatal(err)
	}
	issue.PrintText("")
}

func TestNewJiraIssueV2(t *testing.T) {
	issue, err := NewJiraIssueV2(context.TODO(), jira, issueKey)
	if err != nil {
		t.Fatal(err)
	}
	issue.PrintText("")
}
