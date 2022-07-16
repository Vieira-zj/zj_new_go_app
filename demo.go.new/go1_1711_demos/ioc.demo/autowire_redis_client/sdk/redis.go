package sdk

import (
	"time"

	"github.com/go-redis/redis"
)

// +ioc:autowire=true
// +ioc:autowire:type=normal
// +ioc:autowire:type=singleton
// +ioc:autowire:paramType=Param
// +ioc:autowire:constructFunc=New

type Redis struct {
	client *redis.Client
}

func (r *Redis) Get(key string) *redis.StringCmd {
	return r.client.Get(key)
}

func (r *Redis) Set(key string, value interface{}, expiration time.Duration) *redis.StatusCmd {
	return r.client.Set(key, value, expiration)
}
