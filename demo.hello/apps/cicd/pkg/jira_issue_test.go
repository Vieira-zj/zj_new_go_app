package pkg

import (
	"context"
	"testing"
)

func TestNewJiraIssue(t *testing.T) {
	issue, err := NewJiraIssue(context.TODO(), jira, "AIRPAY-66085")
	if err != nil {
		t.Fatal(err)
	}
	issue.PrintText("")
}
