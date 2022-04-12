package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestAcquireBeforeRateLimiterStart(t *testing.T) {
	const (
		duration  = 3 * time.Second
		threshold = 1
		burst     = 3
	)
	limiter := NewRateLimiter(duration, threshold, burst)
	limiter.Acquire()
}

func TestNewRateLimiter(t *testing.T) {
	const (
		duration  = 2 * time.Second
		threshold = 1
		burst     = 3
	)
	limiter := NewRateLimiter(duration, threshold, burst)
	limiter.Start()
	defer limiter.Stop()

	time.Sleep(5 * time.Second)
	fmt.Println("ratelimiter test done")
}

func TestRateLimiterAcquireWithBlocked(t *testing.T) {
	const (
		duration  = 3 * time.Second
		threshold = 1
		burst     = 3
	)
	limiter := NewRateLimiter(duration, threshold, burst)
	limiter.Start()
	defer limiter.Stop()

	for i := 0; i < 6; i++ {
		if limiter.AcquireWithBlocked(time.Second) {
			fmt.Println("got and run")
		} else {
			fmt.Println("wait second, do not got because of exceed ratelimit")
		}
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Println("ratelimiter test done")
}

// run: go test -timeout 40s -run ^TestRateLimiter$ demo.hello/utils -v -count=1
func TestRateLimiterAcquire(t *testing.T) {
	const (
		duration  = 3 * time.Second
		threshold = 1
		burst     = 3
	)
	limiter := NewRateLimiter(duration, threshold, burst)
	limiter.State()

	myPrint := func() {
		if !limiter.Acquire() {
			fmt.Println("exceed rate limit")
			return
		}
		fmt.Println("foo")
	}

	limiter.Start()
	defer func() {
		limiter.Stop()
		limiter.State()
	}()
	limiter.State()

	for i := 0; i < 10; i++ {
		myPrint()
		time.Sleep(time.Second)
	}
	limiter.State()

	fmt.Println("\nwait...")
	time.Sleep(time.Duration(10) * time.Second)
	limiter.State()

	for i := 0; i < 10; i++ {
		myPrint()
		time.Sleep(time.Second)
	}
	fmt.Println("ratelimiter test done")
}
