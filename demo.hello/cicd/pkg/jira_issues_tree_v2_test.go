package pkg

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestTicketsTreeV2(t *testing.T) {
	tickets := "AIRPAY-66492,SPPAY-2210,AIRPAY-57284,AIRPAY-62043"
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
	defer cancel()

	tree := NewJiraIssuesTreeV2(ctx, 5)
	for _, ticket := range strings.Split(tickets, ",") {
		tree.SubmitIssue(ticket)
	}

	tree.WaitDone()
	fmt.Println(GetIssuesTreeText(tree))
	fmt.Println(GetIssuesTreeUsageText(tree))
}

func TestPrintFixVersionTreeV2(t *testing.T) {
	// fix version -> pm/story tasks -> tasks
	key := "apa_v1.0.20.20210426"
	jql := fmt.Sprintf("fixVersion = %s", key)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
	defer cancel()
	keys, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		t.Fatal(err)
	}

	tree := NewJiraIssuesTreeV2(ctx, 6)
	for _, key := range keys {
		tree.SubmitIssue(key)
	}

	tree.WaitDone()
	fmt.Println(GetIssuesTreeText(tree))
	fmt.Println(GetIssuesTreeUsageText(tree))
}

func TestPrintReleaseCycleTreeV2(t *testing.T) {
	jql := `"Release Cycle" = "2021.04.v4 - AirPay"`
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
	defer cancel()
	keys, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		t.Fatal(err)
	}

	tree := NewJiraIssuesTreeV2(ctx, 6)
	for _, key := range keys {
		tree.SubmitIssue(key)
	}

	tree.WaitDone()
	fmt.Println(GetIssuesTreeText(tree))
	fmt.Println(GetIssuesTreeUsageText(tree))
	fmt.Println(GetIssuesTreeSummary(tree))
}
