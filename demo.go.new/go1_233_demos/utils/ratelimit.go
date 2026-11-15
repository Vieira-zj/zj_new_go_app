package utils

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/go-redis/redis/v8"
)

type Debounce struct {
	delay   time.Duration
	eventCh <-chan struct{}
	fn      func()
	cancel  context.CancelFunc
}

func NewDebounce(ctx context.Context, delay time.Duration, eventCh <-chan struct{}, fn func()) (Debounce, error) {
	if delay < 0 {
		return Debounce{}, fmt.Errorf("delay must be positive: %s", delay)
	}

	ctx, cancel := context.WithCancel(ctx)

	debounce := Debounce{
		delay:   delay,
		eventCh: eventCh,
		fn:      fn,
		cancel:  cancel,
	}
	go debounce.loop(ctx)
	return debounce, nil
}

func (d Debounce) Stop() {
	d.cancel()
}

func (d Debounce) loop(ctx context.Context) {
	timer := time.NewTimer(d.delay)
	timer.Stop()
	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println(ctx.Err())
			return
		case _, ok := <-d.eventCh:
			if !ok {
				return
			}
			// Go 1.22 的 Reset 文档要求, 重置之前, 定时器应当已经停止或已经到期, 并且 time.C channel 已经 pop 为空
			// if !timer.Stop() {
			// 	<-timer.C
			// }
			timer.Reset(d.delay)
		case <-timer.C:
			d.fn()
		}
	}
}

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
