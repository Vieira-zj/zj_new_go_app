package demos

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"
)

// Context

func TestSubCtxTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if ti, ok := ctx.Deadline(); ok {
		t.Log("ctx dead time:", time.Until(ti).Seconds())
	}

	ch := make(chan struct{})
	go func() {
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if ti, ok := ctx.Deadline(); ok {
			t.Log("sub ctx dead time:", time.Until(ti).Seconds())
		}

		t.Log("goroutine wait")
		<-ctx.Done()

		t.Log("goroutine finish")
		close(ch)
	}()

	<-ch
	t.Log("main finish")
}

// Goroutine

func TestGoroutinesWait(t *testing.T) {
	limit := 3
	wg, lwg := sync.WaitGroup{}, sync.WaitGroup{}
	wg.Add(limit)
	lwg.Add(limit)

	for i := range limit {
		go func(idx int) {
			defer wg.Done()
			log.Printf("goroutine %d start", idx)
			func() {
				defer lwg.Done()
				time.Sleep(time.Duration(i) * time.Second)
			}()

			log.Printf("goroutine %d wait", idx)
			lwg.Wait() // all goroutines run at the same time

			log.Printf("goroutine %d finish", idx)
		}(i)
	}

	wg.Wait()
	t.Log("main finish")
}

// Channel

func TestChannel(t *testing.T) {
	t.Run("close ch from sender", func(t *testing.T) {
		ch := make(chan int, 1)
		go func() {
			for num := range ch {
				fmt.Println("receive:", num)
			}
		}()

		for i := range 10 {
			ch <- i
		}
		close(ch)

		time.Sleep(100 * time.Millisecond) // wait go routine finish
		t.Log("finished")
	})
}
