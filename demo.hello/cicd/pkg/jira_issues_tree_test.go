package pkg

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestTicketsTree(t *testing.T) {
	tickets := "AIRPAY-66492,SPPAY-2210,AIRPAY-57284,AIRPAY-62043"
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
	defer cancel()

	tree := NewJiraIssuesTree(ctx, 3)
	for _, ticket := range strings.Split(tickets, ",") {
		if err := tree.SubmitIssue(ticket); err != nil {
			t.Fatal(err)
		}
	}

	for tree.IsRunning() {
		time.Sleep(time.Second)
	}
	fmt.Println(tree.ToText())
	tree.PrintUsage()
}

func TestPrintReleaseTicketTree(t *testing.T) {
	// release ticket -> tasks
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
	defer cancel()
	releaseTicket, err := NewJiraIssue(ctx, jira, "AIRPAY-66425")
	if err != nil {
		t.Fatal(err)
	}

	tree := NewJiraIssuesTree(ctx, 3)
	for _, issueID := range releaseTicket.SubIssues {
		if err := tree.SubmitIssue(issueID); err != nil {
			t.Fatal(err)
		}
	}

	for tree.IsRunning() {
		time.Sleep(time.Second)
	}
	fmt.Println(tree.ToText())
	tree.PrintUsage()
}

func TestPrintFixVersionTree(t *testing.T) {
	// fix version -> pm/story tasks -> tasks
	key := "apa_v1.0.20.20210426"
	jql := fmt.Sprintf("fixVersion = %s", key)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
	defer cancel()
	keys, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		t.Fatal(err)
	}

	tree := NewJiraIssuesTree(ctx, 6)
	for _, key := range keys {
		if err := tree.SubmitIssue(key); err != nil {
			t.Fatal(err)
		}
	}

	for tree.IsRunning() {
		time.Sleep(time.Second)
	}
	fmt.Println(tree.ToText())
	tree.PrintUsage()
}

func TestPrintReleaseCycleTree(t *testing.T) {
	jql := `"Release Cycle" = "2021.04.v3 - AirPay"`
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
	defer cancel()
	keys, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		t.Fatal(err)
	}

	tree := NewJiraIssuesTree(ctx, 6)
	for _, key := range keys {
		if err := tree.SubmitIssue(key); err != nil {
			t.Fatal(err)
		}
	}

	for tree.IsRunning() {
		time.Sleep(time.Second)
	}
	fmt.Println(tree.ToText())
	tree.PrintUsage()
}
