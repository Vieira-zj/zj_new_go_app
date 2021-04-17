package pkg

import (
	"context"
	"testing"
	"time"
)

func TestCreatePMTaskTree(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	releaseTicket, err := NewJiraIssue(context.TODO(), jira, "AIRPAY-66425")
	if err != nil {
		t.Fatal(err)
	}

	worker := NewJiraTaskWorker(ctx, 3)
	worker.Run()

	for _, issueID := range releaseTicket.SubIssues {
		worker.Submit(issueID)
	}

	time.Sleep(time.Duration(5) * time.Second)
	worker.Cache.PrintTaskTree()
}
