package demos

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/errgroup"
)

// Demo: Goroutine

func TestIteratorForChan(t *testing.T) {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, n := range []int{1, 2, 3, 4, 5} {
			ch <- n
		}
	}()

	t.Run("iterator chan by check ok", func(t *testing.T) {
		for {
			if v, ok := <-ch; ok {
				t.Log("value:", v)
			} else {
				t.Log("close")
				break
			}
		}
	})

	t.Run("directly iterator chan", func(t *testing.T) {
		for v := range ch {
			t.Log("value:", v)
		}
		t.Log("done")
	})
}

func TestParallelByLimit(t *testing.T) {
	wg := sync.WaitGroup{}
	limit := make(chan struct{}, 3)

	for i := 0; i < 10; i++ {
		limit <- struct{}{}
		wg.Add(1)
		go func(idx int) {
			fmt.Printf("goroutine-%d run\n", idx)
			time.Sleep(3 * time.Second)
			<-limit
			wg.Done()
		}(i)
	}

	wg.Wait()
	close(limit)
	t.Log("all goroutine done")
}

func TestRecoverForPanic01(t *testing.T) {
	ch := make(chan struct{})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				// print stack from recover
				fmt.Println("recover err:", r)
				fmt.Printf("stack:\n%s", debug.Stack())
			}
			close(ch)
		}()
		fmt.Println("goroutine run...")
		time.Sleep(time.Second)
		panic("mock panic")
	}()

	fmt.Println("wait...")
	<-ch
	fmt.Println("recover demo done")
}

func TestRecoverForPanic02(t *testing.T) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("recover err:", r)
				fmt.Printf("stack:\n%s", debug.Stack())
			}
			fmt.Println("goroutine end")
		}()

		fmt.Println("goroutine start")

		go func() {
			defer func() {
				// sub groutine panic can only recover here, but not in parent groutine
				if r := recover(); r != nil {
					fmt.Println("sub recover err:", r)
					fmt.Printf("stack:\n%s", debug.Stack())
				}
				fmt.Println("sub goroutine end")
			}()

			fmt.Println("sub goroutine start")
			time.Sleep(time.Second)
			panic("mock panic")
		}()
	}()

	time.Sleep(3 * time.Second)
	fmt.Println("recover demo done")
}

// ErrGroup

func TestParallelByErrGroup(t *testing.T) {
	urls := []string{
		"http://www.golang.org/",
		"http://www.google.com/",
		"http://www.somestupidname.com/",
	}

	t.Run("run by errgroup", func(t *testing.T) {
		var g errgroup.Group
		for _, url := range urls {
			url := url
			g.Go(func() error {
				resp, err := http.Get(url)
				if err != nil {
					return err
				}
				defer resp.Body.Close()
				fmt.Printf("fetch url %s status %s\n", url, resp.Status)
				return nil
			})
		}

		err := g.Wait()
		assert.NoError(t, err)
		t.Log("done")
	})

	t.Run("run by errgroup with ctx", func(t *testing.T) {
		var results sync.Map
		// 在任意一个 goroutine 返回错误时, 立即取消其他正在运行的 goroutine
		g, ctx := errgroup.WithContext(context.TODO())

		for _, url := range urls {
			url := url
			g.Go(func() error {
				if url == "http://www.somestupidname.com/" {
					time.Sleep(time.Second)
				}
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil) // use global ctx
				if err != nil {
					return err
				}

				resp, err := http.DefaultClient.Do(req)
				if err != nil {
					return err
				}
				defer resp.Body.Close()

				results.Store(url, resp.Status)
				return nil
			})
		}

		// 等待所有 goroutine 完成并返回第一个错误
		if err := g.Wait(); err != nil {
			t.Log("run error:", err)
		}

		results.Range(func(key, value any) bool {
			fmt.Printf("fetch url %s status %s\n", key, value)
			return true
		})
		t.Log("done")
	})
}

func TestRunLimitByErrGroup(t *testing.T) {
	var g errgroup.Group
	g.SetLimit(3)

	for i := 0; i < 5; i++ {
		result := g.TryGo(func() error {
			t.Logf("goroutine %d is starting", i)
			time.Sleep(time.Second)
			t.Logf("goroutine %d is done", i)
			return nil
		})
		if result {
			t.Logf("goroutine %d started", i)
		} else {
			t.Logf("goroutine %d could not start (limit reached)", i)
		}
	}

	err := g.Wait()
	assert.NoError(t, err)
	t.Log("done")
}
