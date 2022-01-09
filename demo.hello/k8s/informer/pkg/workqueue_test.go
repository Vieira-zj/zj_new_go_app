package pkg

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"
	"time"

	"k8s.io/client-go/util/workqueue"
)

func TestMarshalIndent(t *testing.T) {
	data := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{
		Name: "Foo",
		Age:  31,
	}

	b, err := json.Marshal(&data)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))
	fmt.Println()

	b, err = json.MarshalIndent(&data, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(b))
}

func TestWorkqueue(t *testing.T) {
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "queue-test")
	defer queue.ShutDown()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func(ctx context.Context) {
		i := 0
		for {
			select {
			case <-ctx.Done():
				fmt.Println("worker exit:", ctx.Err())
				queue.ShutDown()
				return
			default:
				i++
				queue.Add(i)
			}
			time.Sleep(300 * time.Millisecond)
		}
	}(ctx)

	go func() {
		for {
			exit := func() bool {
				val, quit := queue.Get()
				if quit {
					fmt.Println("consumer exit")
					return true
				}
				defer queue.Done(val)
				fmt.Printf("get value: %v\n", val)
				return false
			}()
			if exit {
				return
			}
		}
	}()

	<-ctx.Done()
	time.Sleep(200 * time.Millisecond)
	fmt.Println("done")
}

func TestWorkqueueRatelimiter01(t *testing.T) {
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ratelimit-queue-test")
	defer queue.ShutDown()

	const key = "test_key"
	for i := 0; i < 3; i++ {
		queue.AddRateLimited(key)
		time.Sleep(100 * time.Millisecond)
		retKey, _ := queue.Get()
		fmt.Println("get key:", retKey)
		queue.Done(key) // finish handle key
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Printf("length: %d, requeue: %d\n", queue.Len(), queue.NumRequeues(key))
	queue.Forget(key)
	fmt.Printf("length: %d, requeue: %d\n", queue.Len(), queue.NumRequeues(key))

	// add key but not get
	for i := 0; i < 3; i++ {
		queue.AddRateLimited(key)
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Printf("length: %d, requeue: %d\n", queue.Len(), queue.NumRequeues(key))
}

func TestWorkqueueRatelimiter02(t *testing.T) {
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ratelimit-queue-test")
	defer queue.ShutDown()

	key := "re_enqueue"
	closeCh := make(chan struct{})
	go func() {
		for i := 0; i < 6; i++ {
			queue.AddRateLimited(key)
			time.Sleep(200 * time.Millisecond)
		}
		fmt.Println("length of queue:", queue.Len())
		closeCh <- struct{}{}
	}()

	go func() {
		retries := 3
		run := func() bool {
			key, quit := queue.Get()
			if quit {
				fmt.Println("queue quit")
				return false
			}
			defer queue.Done(key)

			if queue.NumRequeues(key) > retries {
				fmt.Println("exceed max retries:", retries)
				queue.Forget(key)
			} else {
				fmt.Println("get value:", key)
			}
			return true
		}

		for run() {
			fmt.Println("number of requeue:", queue.NumRequeues(key))
		}
	}()

	<-closeCh
	queue.ShutDown()
	time.Sleep(200 * time.Millisecond)
	fmt.Println("done")
}

func TestWorkqueueRatelimiter03(t *testing.T) {
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "ratelimit-queue-test")
	defer queue.ShutDown()

	const maxRetries = 3
	handler := func(key string) error {
		retries := queue.NumRequeues(key)
		if retries > maxRetries {
			fmt.Println("number of requeue:", retries)
			return fmt.Errorf("exceed max retries: %d", maxRetries)
		}
		fmt.Println("get value:", key)
		return fmt.Errorf("mock")
	}

	const key = "1"
	queue.Add(key)
	run := func(key string) bool {
		retKey, quit := queue.Get()
		if quit {
			return false
		}
		defer func() {
			// fmt.Println("done for key:", retKey)
			queue.Done(retKey)
		}()

		if err := handler(retKey.(string)); err != nil {
			if err.Error() == "mock" {
				fmt.Println("mock error, and re-queue:", key)
				// AddRateLimited() -> Done()
				queue.AddRateLimited(key)
				return true
			}
			fmt.Println("error:", err)
			fmt.Println("forget key:", key)
			// Forget() -> Done()
			queue.Forget(key)
			return false
		}

		// does not go here
		queue.Forget(key)
		return true
	}

	for run(key) {
		time.Sleep(100 * time.Millisecond)
	}
	fmt.Println("done, number of requeue:", queue.NumRequeues(key))
}

func TestWorkqueueDelay(t *testing.T) {
	queue := workqueue.NewNamedDelayingQueue("delay-queue-test")
	defer queue.ShutDown()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	go func() {
		val := "repeat"
		for i := 1; i <= 3; i++ {
			queue.AddAfter(val, time.Duration(i*150)*time.Millisecond)
		}
		fmt.Println("repeat worker exit")
	}()

	go func() {
		for i := 1; i <= 5; i++ {
			queue.AddAfter(strconv.Itoa(i), time.Duration(i*100)*time.Millisecond)
		}
		fmt.Println("worker exit")
	}()

	go func() {
		start := time.Now()
		for {
			val, quit := queue.Get()
			if quit {
				fmt.Println("consumer exit")
				return
			}
			fmt.Printf("[%d ms] get value: %v\n", time.Since(start)/time.Millisecond, val)
			queue.Done(val)
		}
	}()

	<-ctx.Done()
	queue.ShutDown()
	time.Sleep(200 * time.Millisecond)
	fmt.Println("done")
}
