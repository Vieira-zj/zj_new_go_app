package run

import (
	"context"
	"fmt"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestLimiterWait(t *testing.T) {
	limiter := rate.NewLimiter(3, 5)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; ; i++ {
		fmt.Printf("%03d %s\n", i, time.Now().Format("2006-01-02 15:04:05.000"))
		if err := limiter.Wait(ctx); err != nil {
			fmt.Printf("error: %s\n", err.Error())
			return
		}
		// mock processing
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
}

func TestLimiterReserve(t *testing.T) {
	limiter := rate.NewLimiter(3, 5)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; ; i++ {
		fmt.Printf("%03d %s\n", i, time.Now().Format("2006-01-02 15:04:05.000"))
		reserve := limiter.Reserve()
		if !reserve.OK() {
			fmt.Println("Not allowed to act! Did you remember to set lim.burst to be > 0 ?")
			return
		}

		delayD := reserve.Delay()
		time.Sleep(delayD)
		select {
		case <-ctx.Done():
			fmt.Println("timeout, quit")
			return
		default:
		}
		// mock processing
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
}
