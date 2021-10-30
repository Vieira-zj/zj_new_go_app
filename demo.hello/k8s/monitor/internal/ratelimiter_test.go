package internal

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	limiter := NewRateLimiter(3, 3)
	defer func() {
		limiter.Close()
		time.Sleep(time.Second)
	}()

	myPrint := func() {
		if !limiter.Add("myPrint") {
			fmt.Println("exceed rate limit.")
			return
		}
		fmt.Println("foo")
	}

	go func() {
		limiter.Run(context.Background())
	}()

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
