package pkg

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestOneTicketTree(t *testing.T) {
	ticket := "SPPAY-1236"
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()

	tree := NewJiraIssuesTree(ctx, "Ticket:"+ticket, 1)
	tree.Collect()
	tree.SubmitIssue(ticket)

	for tree.QueueSize() > 0 {
		time.Sleep(time.Second)
	}
	tree.PrintText()
}

func TestPrintReleaseTicketTree(t *testing.T) {
	// release ticket -> tasks
	issueKey := "AIRPAY-66425"
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()
	releaseTicket, err := NewJiraIssue(ctx, jira, issueKey)
	if err != nil {
		t.Fatal(err)
	}

	tree := NewJiraIssuesTree(ctx, "ReleaseTicket:"+issueKey, 3)
	tree.Collect()
	for _, issueID := range releaseTicket.SubIssues {
		tree.SubmitIssue(issueID)
	}

	for tree.QueueSize() > 0 {
		time.Sleep(time.Second)
	}
	tree.PrintText()
	tree.PrintUsage()
}

func TestPrintFixVersionTree(t *testing.T) {
	// fix version -> pm/story tasks -> tasks
	key := "apa_v1.0.19.20210419"
	jql := fmt.Sprintf("fixVersion = %s", key)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
	defer cancel()
	keys, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		t.Fatal(err)
	}

	tree := NewJiraIssuesTree(ctx, "FixVersion:"+key, 6)
	tree.Collect()
	for _, key := range keys {
		tree.SubmitIssue(key)
	}

	for tree.QueueSize() > 0 {
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)
	tree.PrintText()
	tree.PrintUsage()
}

func TestPrintReleaseCycleTree(t *testing.T) {
	key := "2021.04.v3 - AirPay"
	jql := `"Release Cycle" = "2021.04.v3 - AirPay"`
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
	defer cancel()
	keys, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		t.Fatal(err)
	}

	tree := NewJiraIssuesTree(ctx, "ReleaseCycle:"+key, 6)
	tree.Collect()
	for _, key := range keys {
		tree.SubmitIssue(key)
	}

	for tree.QueueSize() > 0 {
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)
	tree.PrintText()
	tree.PrintUsage()
}
