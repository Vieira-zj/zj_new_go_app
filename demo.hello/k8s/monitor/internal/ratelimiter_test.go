package internal

import (
	"fmt"
	"testing"
	"time"
)

// run: go test -timeout 33s -run ^TestRateLimiter$ demo.hello/k8s/monitor/internal -v -count=1
func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(3, 3)
	myPrint := func() {
		if !limiter.Acquire("myPrint") {
			fmt.Println("exceed rate limit")
			return
		}
		fmt.Println("foo")
	}

	limiter.Start()
	defer limiter.Stop()

	for i := 0; i < 10; i++ {
		myPrint()
		time.Sleep(time.Second)
	}

	fmt.Println("\nwait...")
	time.Sleep(time.Duration(10) * time.Second)
	for i := 0; i < 10; i++ {
		myPrint()
		time.Sleep(time.Second)
	}
	fmt.Println("done")
}
