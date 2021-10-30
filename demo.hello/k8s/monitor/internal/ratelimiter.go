package internal

import (
	"context"
	"log"
	"time"
)

// RateLimiter 非并发安全。
type RateLimiter struct {
	duration int
	rate     int
	burst    int
	store    map[string]int
	closeCh  chan struct{}
}

// NewRateLimiter creates rate limiter with duration (seconds), rate (const 1), and burst (max capacity).
func NewRateLimiter(duration int, burst int) *RateLimiter {
	return &RateLimiter{
		duration: duration,
		rate:     1,
		burst:    burst,
		store:    make(map[string]int, 16),
		closeCh:  make(chan struct{}, 1),
	}
}

// Run .
func (rl *RateLimiter) Run(ctx context.Context) {
	tick := time.Tick(time.Duration(rl.duration) * time.Second)
	for {
		select {
		case <-tick:
			for key := range rl.store {
				if rl.store[key] > 0 {
					rl.store[key] -= rl.rate
				}
				if rl.store[key] <= 0 {
					delete(rl.store, key)
				}
			}
		case <-rl.closeCh:
			log.Println("ratelimiter exit by closed.")
			return
		case <-ctx.Done():
			log.Println("ratelimiter exit by cancel.")
			return
		}
	}
}

// Add .
func (rl *RateLimiter) Add(key string) bool {
	var (
		count int
		ok    bool
	)
	if count, ok = rl.store[key]; !ok {
		rl.store[key] = 1
		return true
	}
	if count < rl.burst {
		rl.store[key]++
		return true
	}
	return false
}

// Close .
func (rl *RateLimiter) Close() {
	rl.closeCh <- struct{}{}
}
