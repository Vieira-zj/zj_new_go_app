package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
)

var redisClient *redis.Client

func GetRedisClient(addr, uname, pwd string) *redis.Client {
	if redisClient == nil {
		opt := redis.Options{
			Addr: addr,
		}
		if len(uname) > 0 {
			opt.Username = uname
		}
		if len(pwd) > 0 {
			opt.Password = pwd
		}
		redisClient = redis.NewClient(&opt)
	}

	return redisClient
}

func GetRedisClientForLocalTest(t *testing.T) (*redis.Client, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	client := GetRedisClient("127.0.0.1:6379", "", "")
	err := client.Ping(ctx).Err()
	return client, err
}

// kv

func Add(client *redis.Client, key string, value any, expr time.Duration) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()
	return client.SetNX(ctx, key, value, expr).Err()
}

func GetInt(client *redis.Client, key string) (int, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()
	return client.Get(ctx, key).Int()
}

func Del(client *redis.Client, key string) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	return client.Del(ctx, key).Err()
}

// hash

func HIncrBy(client *redis.Client, key, field string, incr int64) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()
	return client.HIncrBy(ctx, key, field, incr).Err()
}

func HGetInt(client *redis.Client, key, field string) (int, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()
	return client.HGet(ctx, key, field).Int()
}

func HGetAll(client *redis.Client, key string) (map[string]string, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	results, err := client.HGetAll(ctx, key).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		return nil, err
	}
	return results, nil
}

func ReplaceHashFieldName(client *redis.Client, key, oldField, newField string, value any) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	p := client.Pipeline()
	p.HSetNX(ctx, key, newField, value)
	p.HDel(ctx, key, oldField)

	_, err := p.Exec(ctx)
	return err
}

// sorted set (with scores)

func ZIncrBy(client *redis.Client, key, member string, increment float64) error {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()
	return client.ZIncrBy(ctx, key, increment, member).Err()
}

func ZRangeWithScores(client *redis.Client, key string, start, stop int64) ([]redis.Z, error) {
	ctx, cancel := context.WithTimeout(context.TODO(), 3*time.Second)
	defer cancel()

	results, err := client.ZRangeWithScores(ctx, key, start, stop).Result()
	if err != nil && errors.Is(err, redis.Nil) {
		return nil, err
	}
	return results, nil
}
