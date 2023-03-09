package redis

import (
	"context"
	"sync"
	"testing"
	"time"
)

// run: go test -timeout 180s -run ^TestRedisLock$ go1_1711_demo/utils -v -count=1
func TestRedisLock(t *testing.T) {
	client := getRedisClientInitForLocal()
	key := "redis.test"
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		const tag = "goroutine1-1"
		lock := NewRedisLock(client, key, tag)
		ctx, cancel := context.WithCancel(context.Background())

		defer func() {
			cancel()
			res, err := lock.Release(context.Background())
			if err != nil {
				t.Logf("[%s] release lock error: %v", tag, err)
			} else if !res {
				t.Logf("[%s] release lock failed", tag)
			} else {
				t.Logf("[%s] release lock: %s", tag, lock.String())
			}
			wg.Done()
		}()

		// acquire lock
		res, err := lock.Acquire(context.Background(), 15)
		if err != nil {
			t.Logf("[%s] acquire lock error: %v", tag, err)
			return
		}
		if !res {
			t.Logf("[%s] acquire lock fail", tag)
			return
		}
		t.Logf("[%s] acquire lock", tag)

		// goroutine to extend lock. if extend lock failed?
		go func() {
			const tag = "goroutine1-2"
			tick := time.Tick(10 * time.Second)
			lctx := context.Background()
			for i := 0; i < 10; i++ {
				select {
				case <-ctx.Done():
					t.Logf("[%s] extend lock is cancelled", tag)
					return
				case <-tick:
					t.Logf("[%s] extend lock", tag)
					res, err := lock.Acquire(lctx, 15)
					if err != nil {
						t.Logf("[%s] extend lock error: %v", tag, err)
					}
					if !res {
						t.Logf("[%s] extend lock fail", tag)
					}
				}
			}
		}()

		t.Logf("[%s] running ...", tag)
		time.Sleep(32 * time.Second)
		t.Logf("[%s] finish", tag)
	}()

	time.Sleep(200 * time.Millisecond)
	wg.Add(1)
	go func() {
		const tag = "goroutine2"
		lock := NewRedisLock(client, key, tag)
		defer func() {
			res, err := lock.Release(context.Background())
			if err != nil {
				t.Logf("[%s] release lock error: %v", tag, err)
			} else if !res {
				t.Logf("[%s] release lock failed", tag)
			} else {
				t.Logf("[%s] release lock: %s", tag, lock.String())
			}
			wg.Done()
		}()

		ctx := context.Background()
		for i := 1; i < 45; i++ {
			res, err := lock.Acquire(ctx, 15)
			if err == nil && res {
				t.Logf("[%s] acuqire lock", tag)
				time.Sleep(3 * time.Second)
				t.Logf("[%s] finish", tag)
				return
			}
			if err != nil {
				t.Logf("[%s] acquire lock error: %v", tag, err)
			} else if !res {
				t.Logf("[%s] acquire lock failed, wait and retry", tag)
			}
			time.Sleep(time.Second)
		}
	}()

	wg.Wait()
	t.Log("redis lock done")
}
