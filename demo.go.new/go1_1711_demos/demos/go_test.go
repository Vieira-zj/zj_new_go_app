package demos

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestChar(t *testing.T) {
	c := fmt.Sprintf("%c", 119)
	t.Logf("str=%s, len=%d", c, len(c))

	c = fmt.Sprintf("%c", 258)
	t.Logf("str=%s, len=%d", c, len(c))

	r := rune('中')
	t.Logf("char=%c, d=%d", r, r)
	c = fmt.Sprintf("%c", r)
	t.Logf("str=%s, len=%d", c, len(c))

	c = fmt.Sprintf("%c", 20132)
	t.Logf("str=%s, len=%d", c, len(c))

	s := "中cn"
	t.Logf("size=%d", len(s))
}

func TestMarshalFunc(t *testing.T) {
	// json.Marshal unsupported type: func()
	type caller struct {
		Name string `json:"name"`
		Fn   func() `json:"func"`
	}

	c := &caller{
		Name: "helloworld",
		Fn: func() {
			fmt.Println("helloworld")
		},
	}
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("caller: %s\n", b)
}

func TestRunBatchByGoroutine(t *testing.T) {
	resultCh := make(chan int)
	errCh := make(chan error)
	defer func() {
		close(resultCh)
		close(errCh)
	}()

	go func(resultCh chan int, errCh chan error) {
		for i := 0; i < 10; i++ {
			if i == 11 {
				errCh <- fmt.Errorf("invalid num")
			}
			resultCh <- i
			time.Sleep(time.Second)
		}
		// NOTE: size of resultCh should be 0, make sure all results are handle before return
		errCh <- nil
	}(resultCh, errCh)

outer:
	for {
		select {
		case result := <-resultCh:
			if result%2 == 1 {
				continue
			}
			t.Log("result:", result)
		case err := <-errCh:
			if err != nil {
				t.Log("err:", err)
			}
			break outer
		}
	}
	t.Log("done")
}

func TestGoroutineExit(t *testing.T) {
	// NOTE: sub goroutine is still running when root goroutine exit
	// ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer cancel()

	retCh := make(chan struct{})
	go func() {
		// context here, make sure sub goroutine is cancelled when root goroutine exit
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		go func() {
			tick := time.Tick(time.Second)
			for {
				select {
				case <-ctx.Done():
					fmt.Println("sub goroutine:", ctx.Err())
					return
				case <-tick:
					fmt.Println("sub goroutine run...")
				}
			}
		}()
		for i := 0; i < 5; i++ {
			time.Sleep(time.Second)
			fmt.Println("root goroutine run...")
		}

		// <-retCh
		close(retCh)
	}()

	t.Log("main wait...")
	<-retCh
	t.Log("root goroutine finish")
	time.Sleep(3 * time.Second)
	t.Log("main done")
}
