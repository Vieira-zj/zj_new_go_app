package utils

import (
	"fmt"
	"sync"

	"golang.org/x/time/rate"
)

// IPRateLimiter rate limiter by diff ips.
type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

// NewIPRateLimiter creates an instance of rate limiter.
func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

// AddIP creates a new rate limiter for ip.
func (i *IPRateLimiter) AddIP(ip string) *rate.Limiter {
	i.mu.Lock()
	defer i.mu.Unlock()

	fmt.Printf("add a rate limiter for ip: [%s]\n", ip)
	limiter := rate.NewLimiter(i.r, i.b)
	i.ips[ip] = limiter
	return limiter
}

// GetLimiter returns the rate limiter for the provided IP address if it exists, otherwise add.
func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()

	limiter, ok := i.ips[ip]
	if !ok {
		i.mu.RUnlock()
		return i.AddIP(ip)
	}
	i.mu.RUnlock()
	return limiter
}
