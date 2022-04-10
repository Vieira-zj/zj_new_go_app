package utils

import (
	"log"
	"sync/atomic"
	"time"
)

// RateLimiter .
type RateLimiter struct {
	isRunning int32
	duration  time.Duration
	threshold int
	burst     int
	store     chan struct{}
	closeCh   chan struct{}
}

// NewRateLimiter creates rate limiter with rate (threshold/duration), and burst (max capacity).
func NewRateLimiter(duration time.Duration, threshold, burst int) *RateLimiter {
	if threshold > burst {
		log.Fatalln("threshold should be >= burst.")
	}

	return &RateLimiter{
		duration:  duration,
		threshold: threshold,
		burst:     burst,
	}
}

// Start .
func (rl *RateLimiter) Start() {
	if atomic.LoadInt32(&rl.isRunning) > 0 {
		return
	}

	atomic.AddInt32(&rl.isRunning, 1)
	rl.store = make(chan struct{}, rl.burst)
	rl.closeCh = make(chan struct{})

	for i := 0; i < rl.burst; i++ {
		rl.store <- struct{}{}
	}

	go func() {
		tick := time.Tick(rl.duration)
		for {
			select {
			case <-tick:
				for i := 0; i < rl.threshold; i++ {
					select {
					case rl.store <- struct{}{}:
					default:
						log.Println("not fill because of exceed the burst")
					}
				}
			case <-rl.closeCh:
				log.Println("ratelimiter exit")
				return
			}
		}
	}()
}

// Acquire .
func (rl *RateLimiter) Acquire() bool {
	if atomic.LoadInt32(&rl.isRunning) == 0 {
		log.Fatalln("ratelimit is already closed")
	}

	select {
	case <-rl.store:
		return true
	default:
		return false
	}
}

// AcquireWithBlocked .
func (rl *RateLimiter) AcquireWithBlocked(timeout time.Duration) bool {
	if atomic.LoadInt32(&rl.isRunning) == 0 {
		log.Fatalln("ratelimit is already closed")
	}

	select {
	case <-rl.store:
		return true
	case <-time.After(timeout):
		return false
	}
}

// State .
func (rl *RateLimiter) State() {
	log.Printf("ratelimiter state: is_running=%d, current_total_tokens=%d", atomic.LoadInt32(&rl.isRunning), len(rl.store))
}

// Stop .
func (rl *RateLimiter) Stop() {
	if atomic.LoadInt32(&rl.isRunning) == 0 {
		return
	}

	atomic.AddInt32(&rl.isRunning, -1)
	close(rl.store)
	close(rl.closeCh)
}
