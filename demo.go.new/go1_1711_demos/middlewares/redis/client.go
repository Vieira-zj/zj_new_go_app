package redis

import (
	"sync"

	"github.com/go-redis/redis"
)

var (
	redisClient     *redis.Client
	redisClientOnce sync.Once
)

func NewRedisClient(addr string) *redis.Client {
	redisClientOnce.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "",
			DB:       1,
			PoolSize: 10,
		})
	})
	return redisClient
}
