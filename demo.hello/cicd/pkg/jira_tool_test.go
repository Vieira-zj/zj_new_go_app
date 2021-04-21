package pkg

import (
	"context"
	"fmt"
	"testing"
)

var issueID = "AIRPAY-61523"

func TestJiraGetIssue(t *testing.T) {
	resp, err := jira.GetIssue(context.TODO(), issueID, nil)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("get issue:", string(resp))
}

func TestJiraGetIssueByFields(t *testing.T) {
	fields := []string{"key", "summary"}
	resp, err := jira.GetIssue(context.TODO(), issueID, fields)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("get issue:", string(resp))
}

func TestJiraSearch(t *testing.T) {
	jql := `"Release Cycle" = "2021.04.v3" AND type = Bug`
	fields := []string{"key", "summary", "status"}
	resp, err := jira.Search(context.TODO(), jql, fields)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("search results:", string(resp))
}

func TestJiraSearchIssues(t *testing.T) {
	jql := `fixVersion = apa_v1.0.19.20210419`
	keys, err := jira.SearchIssues(context.TODO(), jql)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(keys)
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

func TestGetIssuesInEpic(t *testing.T) {
	keys, err := jira.GetIssuesInEpic(context.TODO(), "SPPAY-196")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(keys)
}
