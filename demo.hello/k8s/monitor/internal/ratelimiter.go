package internal

import (
	"log"
	"time"
)

// RateLimiter 非并发安全。
type RateLimiter struct {
	duration  int
	threshold int
	burst     int
	store     map[string]int
	closeCh   chan struct{}
}

// NewRateLimiter creates rate limiter with rate (threshold/duration), and burst (max capacity).
func NewRateLimiter(duration, burst int) *RateLimiter {
	const threshold = 1
	if burst < threshold {
		log.Fatalln("burst should be >= threshold (const 1)")
	}

	return &RateLimiter{
		duration:  duration,
		threshold: threshold,
		burst:     burst,
	}
}

// Start .
func (rl *RateLimiter) Start() {
	rl.store = make(map[string]int, 16)
	rl.closeCh = make(chan struct{})

	go func() {
		tick := time.Tick(time.Duration(rl.duration) * time.Second)
		for {
			select {
			case <-tick:
				for key := range rl.store {
					rl.store[key] -= rl.threshold
					if rl.store[key] < 0 {
						delete(rl.store, key)
					}
				}
			case <-rl.closeCh:
				log.Println("ratelimiter exit by closed.")
				return
			}
		}
	}()
}

// Acquire .
func (rl *RateLimiter) Acquire(key string) bool {
	count, ok := rl.store[key]
	if !ok {
		rl.store[key] = 1
		return true
	}
	if count < rl.burst {
		rl.store[key]++
		return true
	}
	return false
}

// Stop .
func (rl *RateLimiter) Stop() {
	close(rl.closeCh)
}
