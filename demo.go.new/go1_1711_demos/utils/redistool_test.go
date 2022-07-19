package utils

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

func RedisLocalTestInit() *redis.Client {
	addr := "127.0.0.1:6379"
	return NewRedisClient(addr)
}

func TestRedisConnPing(t *testing.T) {
	client := RedisLocalTestInit()
	res, err := client.Ping().Result()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ping:", res)
}

func TestRedisConnSet(t *testing.T) {
	client := RedisLocalTestInit()
	const key = "redis.test"
	if _, err := client.Set(key, "valtest", time.Minute).Result(); err != nil {
		t.Fatal(err)
	}
	t.Log("set success")
}

func TestRedisConnGet(t *testing.T) {
	client := RedisLocalTestInit()
	const key = "redis.test"
	res, err := client.Get(key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			t.Log("get success: val nil")
		} else {
			t.Fatal(err)
		}
	} else {
		t.Log("get success:", res)
	}
}

func TestRedisConnDel(t *testing.T) {
	client := RedisLocalTestInit()
	const key = "redis.test"
	if _, err := client.Del(key).Result(); err != nil {
		t.Fatal(err)
	}
	t.Log("del success")
}

// run: go test -timeout 180s -run ^TestRedisLock$ go1_1711_demo/utils -v -count=1
func TestRedisLock(t *testing.T) {
	client := RedisLocalTestInit()
	key := "redis.test"
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		const tag = "goroutine1-1"
		lock := NewRedisLock(client, key, tag)
		ctx, cancel := context.WithCancel(context.Background())

		defer func() {
			cancel()
			res, err := lock.Release()
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
		res, err := lock.Acquire(15)
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
			for i := 0; i < 10; i++ {
				select {
				case <-ctx.Done():
					t.Logf("[%s] extend lock is cancelled", tag)
					return
				case <-tick:
					t.Logf("[%s] extend lock", tag)
					res, err := lock.Acquire(15)
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
			res, err := lock.Release()
			if err != nil {
				t.Logf("[%s] release lock error: %v", tag, err)
			} else if !res {
				t.Logf("[%s] release lock failed", tag)
			} else {
				t.Logf("[%s] release lock: %s", tag, lock.String())
			}
			wg.Done()
		}()

		for i := 1; i < 45; i++ {
			res, err := lock.Acquire(15)
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
