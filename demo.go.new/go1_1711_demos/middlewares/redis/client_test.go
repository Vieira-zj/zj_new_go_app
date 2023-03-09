package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	redis "github.com/go-redis/redis/v8"
)

func getRedisClientInitForLocal() *redis.Client {
	addr := "127.0.0.1:6379"
	return NewRedisClient(addr)
}

func TestRedisClientPing(t *testing.T) {
	client := getRedisClientInitForLocal()
	res, err := client.Ping(context.Background()).Result()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("ping:", res)
}

func TestRedisClientSet(t *testing.T) {
	client := getRedisClientInitForLocal()
	const key = "redis.test"
	if _, err := client.Set(context.Background(), key, "valtest", time.Minute).Result(); err != nil {
		t.Fatal(err)
	}
	t.Log("set success")
}

func TestRedisClientGet(t *testing.T) {
	client := getRedisClientInitForLocal()
	const key = "redis.test"
	res, err := client.Get(context.Background(), key).Result()
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
	client := getRedisClientInitForLocal()
	const key = "redis.test"
	if _, err := client.Del(context.Background(), key).Result(); err != nil {
		t.Fatal(err)
	}
	t.Log("del success")
}

func TestRedisBitMap(t *testing.T) {
	client := getRedisClientInitForLocal()
	const key = "bitmap.testkey"
	pipe := client.TxPipeline()

	// init
	ctx := context.Background()
	pipe.SetBit(ctx, key, 100, 1)
	// append
	for i := 0; i < 60; i++ {
		if i%2 == 0 {
			pipe.SetBit(ctx, key, int64(i), 1)
		}
	}

	if _, err := pipe.Exec(ctx); err != nil {
		pipe.Discard()
		t.Fatal(err)
	}
	if err := pipe.Close(); err != nil {
		t.Fatal(err)
	}

	// check count
	result, err := client.BitCount(ctx, key, &redis.BitCount{Start: 0, End: 100}).Result()
	if err != nil {
		t.Fatal(err)
	}
	t.Log("bit count:", result)

	// check exist
	n, err := client.Exists(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("key [%s] exist: %v", key, n == 1)

	for _, offset := range []int64{2, 21, 54, 101} {
		result := client.GetBit(ctx, key, offset)
		if err := result.Err(); err != nil {
			t.Fatal(err)
		}
		t.Logf("offset=%d, exist=%v", offset, result.Val() == 1)
	}

	if result := client.Del(ctx, key); result.Err() != nil {
		t.Fatal(result.Err())
	}

	// clear
	time.Sleep(time.Second)
	n, err = client.Exists(ctx, key).Result()
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("after clear, key [%s] exist: %v", key, n == 1)
}
