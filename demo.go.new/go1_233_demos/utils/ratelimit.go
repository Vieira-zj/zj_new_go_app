package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

// RateLimit impl by redis, which burst is 60, and fill rate is 20 per second.
type RateLimit struct {
	client   *redis.Client
	key      string
	burst    int
	fillRate int

	maxRetry int
	interval time.Duration
}

func NewRateLimit(client *redis.Client, key string, burst, fillRate int) RateLimit {
	if burst <= 0 {
		burst = 60
	}
	if fillRate <= 0 {
		fillRate = 20
	}

	return RateLimit{
		client:   client,
		key:      key,
		burst:    burst,
		fillRate: fillRate,

		maxRetry: 60,
		interval: time.Second,
	}
}

func (r *RateLimit) AllowWithTimeout(ctx context.Context, seconds int) (bool, error) {
	ctx, cancel := context.WithTimeoutCause(ctx, time.Duration(seconds)*time.Second, fmt.Errorf("timeout exceed"))
	defer cancel()

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for range r.maxRetry {
		ok, err := r.Allow(ctx)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}

		select {
		case <-ctx.Done():
			return false, context.Cause(ctx)
		case <-ticker.C:
		}
	}
	return false, fmt.Errorf("max retry exceed")
}

func (r *RateLimit) Allow(ctx context.Context) (bool, error) {
	script := redis.NewScript(`
local key = KEYS[1]
local burst = tonumber(ARGV[1]) -- max burst capacity  
local fill_rate = tonumber(ARGV[2]) -- tokens per second
local now = tonumber(ARGV[3]) -- current timestamp in ms

local value = redis.call("HMGET", key, "tokens", "last_timestamp")
local last_tokens = tonumber(value[1])
local last_ts = tonumber(value[2])

if last_tokens == nil then
	last_tokens = burst
	last_ts = now
end

local delta = math.max(0, now - last_ts)
tokens = math.min(burst, last_tokens + delta * fill_rate / 1000)

local allowed = tokens >= 1
if allowed then
	tokens = tokens - 1
end

redis.call("HMSET", key, "tokens", tokens, "last_timestamp", now)
redis.call("EXPIRE", key, 86400) -- expire in 24 hours for cleanup

return {allowed, tokens}
`)

	result, err := script.Run(ctx, r.client, []string{r.key}, r.burst, r.fillRate, time.Now().UnixMilli()).Result()
	if err != nil {
		return false, err
	}

	results, ok := result.([]any)
	if !ok && len(results) != 2 {
		return false, nil
	}

	allowed, ok := results[0].(bool)
	if !ok {
		return false, nil
	}
	remaining, ok := results[1].(int64)
	if !ok {
		return false, nil
	}
	log.Printf("allowed=%v, remaining_tokens=%.0f", allowed, float64(remaining))
	return allowed, nil
}
