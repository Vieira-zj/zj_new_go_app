package pkg

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestBlockedQueue(t *testing.T) {
	parallelCount := 5
	queue := make(chan int, 10)
	wg := sync.WaitGroup{}

	for i := 0; i < parallelCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for val := range queue {
				fmt.Printf("go func: %d, time: %d\n", val, time.Now().Unix())
				time.Sleep(time.Second)
			}
		}()
	}

	for i := 0; i < 30; i++ {
		queue <- i
		queue <- i + 1
	}

	fmt.Println("queue size:", len(queue))
	close(queue)
	fmt.Println("queue closed.")
	// 1. 读取完chan中已发送的数据，for range 循环退出；2. 所有goroutine执行完成
	wg.Wait()
}

func TestRemoveDulpicatedItem(t *testing.T) {
	s := []string{"a", "b", "c", "d", "c", "a"}
	ret := removeDulpicatedItem(s)
	fmt.Println(strings.Join(ret, ","))
}

func TestTicketsTree(t *testing.T) {
	tickets := "AIRPAY-66492,SPPAY-2210,AIRPAY-57284,AIRPAY-62043"
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
	defer cancel()

	tree := NewJiraIssuesTree(ctx, 3)
	for _, ticket := range strings.Split(tickets, ",") {
		tree.SubmitIssue(ticket)
	}

	tree.WaitDone()
	fmt.Println(GetIssuesTreeText(tree))
	fmt.Println(GetIssuesTreeUsageText(tree))
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
		tree.SubmitIssue(issueID)
	}

	tree.WaitDone()
	fmt.Println(GetIssuesTreeText(tree))
	fmt.Println(GetIssuesTreeUsageText(tree))
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
		tree.SubmitIssue(key)
	}

	tree.WaitDone()
	fmt.Println(GetIssuesTreeText(tree))
	fmt.Println(GetIssuesTreeUsageText(tree))
}

func TestPrintReleaseCycleTree(t *testing.T) {
	jql := `"Release Cycle" = "2021.04.v4 - AirPay"`
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(20)*time.Second)
	defer cancel()
	keys, err := jira.SearchIssues(ctx, jql)
	if err != nil {
		t.Fatal(err)
	}

	tree := NewJiraIssuesTree(ctx, 6)
	for _, key := range keys {
		tree.SubmitIssue(key)
	}

	tree.WaitDone()
	fmt.Println(GetIssuesTreeText(tree))
	fmt.Println(GetIssuesTreeUsageText(tree))
}
