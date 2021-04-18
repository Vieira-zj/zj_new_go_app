package pkg

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestPrintReleaseTicketTree(t *testing.T) {
	// release ticket -> tasks
	key := "AIRPAY-66425"
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(10)*time.Second)
	defer cancel()
	releaseTicket, err := NewJiraIssue(ctx, jira, key)
	if err != nil {
		t.Fatal(err)
	}

	worker := NewJira2LevelsTreeWorker(ctx, "ReleaseTicket:"+key, 5)
	worker.Start()
	for _, issueID := range releaseTicket.SubIssues {
		worker.Submit(issueID)
	}

	for worker.QueueSize() > 0 {
		time.Sleep(time.Second)
	}
	worker.GetStore().PrintTree()
	worker.GetStore().PrintUsage()
}

func TestPrintFixVersionTree(t *testing.T) {
	// fix version -> pm/story tasks -> tasks
	key := "apa_v1.0.19.20210419"
	jql := fmt.Sprintf("fixVersion = %s", key)
	keys, err := jira.SearchIssues(context.TODO(), jql)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
	defer cancel()
	worker := NewJira2LevelsTreeWorker(ctx, "FixVersion:"+key, 5)
	worker.Start()
	for _, key := range keys {
		worker.Submit(key)
	}

	for worker.QueueSize() > 0 {
		time.Sleep(time.Second)
	}
	time.Sleep(time.Second)
	fmt.Println()
	worker.GetStore().PrintTree()
	worker.GetStore().PrintUsage()
}
