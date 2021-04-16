package utils

import (
	"context"
	"fmt"
	"testing"
)

var (
	jira    = NewJiraTool()
	issueID = "SPPAY-1826"
)

func TestJiraGetIssue(t *testing.T) {
	fields := []string{"key", "summary"}
	resp, err := jira.GetIssue(context.TODO(), issueID, fields)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("get issue:", string(resp))
}

func TestJiraSearch(t *testing.T) {
	jql := `"Release Cycle" = "2021.04.v3" AND type = Bug`
	resp, err := jira.Search(context.TODO(), jql)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("search results:", string(resp))
}

func TestGetIssueLink(t *testing.T) {
	resp, err := jira.GetIssueLink(context.TODO(), issueID)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("issue links:", string(resp))
}

func TestGetRemoteLink(t *testing.T) {
	resp, err := jira.GetRemoteLink(context.TODO(), issueID)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("remote links:", string(resp))
}
