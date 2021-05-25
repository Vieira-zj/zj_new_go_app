package pkg

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestSingleTicketV2(t *testing.T) {
	ticket := "SPPAY-3608"
	tree := NewJiraIssuesTreeV2(2)
	tree.SubmitIssue(ticket)
	tree.WaitDone()
	fmt.Println(GetIssuesTreeText(tree))
}

func TestTicketsTreeV2(t *testing.T) {
	tickets := "AIRPAY-66492,SPPAY-2210,AIRPAY-57284,AIRPAY-62043"
	tree := NewJiraIssuesTreeV2(5)
	for _, ticket := range strings.Split(tickets, ",") {
		tree.SubmitIssue(ticket)
	}

	tree.WaitDone()
	fmt.Println(GetIssuesTreeText(tree))
	fmt.Println(GetIssuesTreeUsageText(tree))
}

func TestPrintFixVersionTreeV2(t *testing.T) {
	// fix version -> pm/story tasks -> tasks
	jira := NewJiraTool()
	key := "apa_v1.0.21.20210430"
	jql := fmt.Sprintf("fixVersion = %s", key)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(8)*time.Second)
	defer cancel()
	keys, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		t.Fatal(err)
	}

	tree := NewJiraIssuesTreeV2(6)
	for _, key := range keys {
		tree.SubmitIssue(key)
	}

	tree.WaitDone()
	fmt.Println(GetIssuesTreeText(tree))
	fmt.Println(GetIssuesTreeUsageText(tree))
}

func TestPrintReleaseCycleTreeV2(t *testing.T) {
	jira := NewJiraTool()
	jql := `"Release Cycle" = "2021.04.v4 - AirPay"`
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
	fmt.Println(GetIssuesTreeUsageText(tree))
	fmt.Println(GetIssuesTreeSummaryText(tree))
}
