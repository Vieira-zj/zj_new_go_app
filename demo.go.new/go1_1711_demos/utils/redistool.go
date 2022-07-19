package utils

import (
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/go-redis/redis"
)

var (
	redisClient        *redis.Client
	NewRedisClientOnce sync.Once
)

func NewRedisClient(addr string) *redis.Client {
	NewRedisClientOnce.Do(func() {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: "",
			DB:       1,
			PoolSize: 10,
		})
	})
	return redisClient
}

// Redis Lock

/*
SET key value [EX seconds] [PX milliseconds] [NX|XX]
- EX second: Sets the expiration time of the key to given seconds. SET key value EX second has the same effect as SETEX key second value.
- PX millisecond : Sets the key expiration time to given milliseconds. SET key value PX millisecond has the same effect as PSETEX key millisecond value.
- NX: Set the key only when the key does not exist. SET key value NX has the same effect as SETNX key value.
- XX: Sets the key only if the key already exists. 2.
*/

/*
- KEYS[1]: lock key
- ARGV[1]: lock value, random string
- ARGV[2]: expiration time

If equal, it means that the lock is acquired again and the acquisition time is updated to prevent expiration on reentry, this means it is a "reentrant lock".
If not, SET key value NX PX timeout: Set the value of the key only when the key does not exist.
Set success will automatically return "OK", set failure returns "NULL Bulk Reply".
*/

const lockCommand = `if redis.call("GET", KEYS[1]) == ARGV[1] then
	redis.call("SET", KEYS[1], ARGV[1], "EX", ARGV[2])
	return "OK"
else
	return redis.call("SET", KEYS[1], ARGV[1], "NX", "EX", ARGV[2])
end`

/*
Release the lock, but cannot release someone else's lock.
*/

const delCommand = `if redis.call("GET", KEYS[1]) == ARGV[1] then
    return redis.call("DEL", KEYS[1])
else
    return 0
end`

type RedisLock struct {
	store *redis.Client
	key   string
	id    string
}

func NewRedisLock(client *redis.Client, key, id string) *RedisLock {
	return &RedisLock{
		store: client,
		key:   key,
		id:    id,
	}
}

func (rl *RedisLock) Acquire(expireSecs int) (bool, error) {
	resp, err := rl.store.Eval(lockCommand, []string{rl.key}, []string{rl.id, strconv.Itoa(expireSecs)}).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil
		}
		return false, err
	}
	if resp == nil {
		return false, nil
	}

	reply, ok := resp.(string)
	if ok && reply == "OK" {
		return true, nil
	}
	return false, fmt.Errorf("unknown reply")
}

func (rl *RedisLock) Release() (bool, error) {
	resp, err := rl.store.Eval(delCommand, []string{rl.key}, []string{rl.id}).Result()
	if err != nil {
		return false, err
	}

	reply, ok := resp.(int64)
	if !ok {
		return false, fmt.Errorf("invalid reply")
	}
	return reply == 1, nil
}

func (rl *RedisLock) String() string {
	return fmt.Sprintf("key=%s,id=%s", rl.key, rl.id)
}
