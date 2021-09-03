package utils

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// limiter init cap with 5, and put 1 token per second.
var ratelimiter = NewIPRateLimiter(1, 5)

func printHello() {
	l := ratelimiter.GetLimiter("127.0.0.1")
	if !l.Allow() {
		fmt.Println("too many access, not allow.")
		return
	}
	fmt.Println("hello world")
}

func TestIPRateLimiter(t *testing.T) {
	wg := sync.WaitGroup{}
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 5; j++ {
				printHello()
				time.Sleep(time.Second)
			}
		}()
	}
	wg.Wait()

	fmt.Println("wait ...")
	time.Sleep(time.Duration(5) * time.Second)
	for i := 0; i < 10; i++ {
		printHello()
	}
	fmt.Println("IPRateLimiter test done.")
}
