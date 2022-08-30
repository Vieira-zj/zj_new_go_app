package redis

import (
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

func RedisClientInitForLocal() *redis.Client {
	addr := "127.0.0.1:6379"
	return NewRedisClient(addr)
}

func TestRedisClientPing(t *testing.T) {
	client := RedisClientInitForLocal()
	res, err := client.Ping().Result()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ping:", res)
}

func TestRedisClientSet(t *testing.T) {
	client := RedisClientInitForLocal()
	const key = "redis.test"
	if _, err := client.Set(key, "valtest", time.Minute).Result(); err != nil {
		t.Fatal(err)
	}
	t.Log("set success")
}

func TestRedisClientGet(t *testing.T) {
	client := RedisClientInitForLocal()
	const key = "redis.test"
	res, err := client.Get(key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			t.Log("get success: val nil")
		} else {
			t.Fatal(err)
		}
	} else {
		t.Log("get success:", res)
	}
}

func TestRedisClientDel(t *testing.T) {
	client := RedisClientInitForLocal()
	const key = "redis.test"
	if _, err := client.Del(key).Result(); err != nil {
		t.Fatal(err)
	}
	t.Log("del success")
}
