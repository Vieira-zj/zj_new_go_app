package redis_test

import (
	"errors"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"demo.apps/middlewares/redis"
	redisv8 "github.com/go-redis/redis/v8"
)

func TestSetExpired(t *testing.T) {
	keyForTest := "test_expired"
	client, err := redis.GetRedisClientForLocalTest(t)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("set expired ok", func(t *testing.T) {
		if err := redis.Add(client, keyForTest, "expired_at_sec", 0); err != nil {
			t.Fatal(err)
		}
		result, err := redis.Get(client, keyForTest)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("get value:", result)

		ok, err := redis.SetExpired(client, keyForTest, time.Second)
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Fatal("add expired failed")
		}
		t.Log("add expired")

		time.Sleep(1200 * time.Millisecond)
		if _, err = redis.Get(client, keyForTest); err != nil && errors.Is(err, redisv8.Nil) {
			t.Logf("key [%s] is expired", keyForTest)
		}
	})

	t.Run("set expired failed", func(t *testing.T) {
		if err := redis.Add(client, keyForTest, "expired_at_3_secs", 3*time.Second); err != nil {
			t.Fatal(err)
		}
		time.Sleep(time.Second)

		_, err := redis.SetExpired(client, keyForTest, time.Second)
		if err != nil && errors.Is(err, redis.ErrExpiredAlreadySet) {
			t.Logf("%v for key: %s", err, keyForTest)
		}

		result, err := redis.Get(client, keyForTest)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("get value:", result)
	})
}

func TestHash(t *testing.T) {
	keyForTest := "test_hash"
	client, err := redis.GetRedisClientForLocalTest(t)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("hash increase by", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			if err := redis.HIncrBy(client, keyForTest, "key_"+strconv.Itoa(i), int64(i)); err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("hash length", func(t *testing.T) {
		count, err := redis.HLen(client, keyForTest)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("hash count:", count)
	})

	t.Run("hash get all", func(t *testing.T) {
		results, err := redis.HGetAll(client, keyForTest)
		if err != nil {
			t.Fatal(err)
		}

		for k, v := range results {
			i, err := strconv.Atoi(v)
			if err != nil {
				t.Fatal(err)
			}
			t.Logf("key=%s, value=%d", k, i)
		}
	})
}

func TestHashHGet(t *testing.T) {
	keyForTest := "test_hash"
	client, err := redis.GetRedisClientForLocalTest(t)
	if err != nil {
		t.Fatal(err)
	}

	redis.HIncrBy(client, keyForTest, "key_sub_1", 1)
	redis.HIncrBy(client, keyForTest, "key_sub_2", 2)

	t.Run("get success", func(t *testing.T) {
		result, ok, err := redis.HGetInt(client, keyForTest, "key_sub_1")
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Log("get:", result)
		}
	})

	t.Run("key not exist", func(t *testing.T) {
		_, ok, err := redis.HGetInt(client, "not_exist", "key_sub_1")
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Log("key not found")
		}
	})

	t.Run("field not exist", func(t *testing.T) {
		_, ok, err := redis.HGetInt(client, keyForTest, "not_exist")
		if err != nil {
			t.Fatal(err)
		}
		if !ok {
			t.Log("field not found")
		}
	})
}

func TestHashHScan(t *testing.T) {
	keyForTest := "test_hash"
	client, err := redis.GetRedisClientForLocalTest(t)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 33; i++ {
		if err := redis.HIncrBy(client, keyForTest, "key_"+strconv.Itoa(i), int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	const limit = int64(10)
	cur := uint64(0)
	for {
		t.Logf("scan: start=%d, end=%d", cur, cur+uint64(limit))
		results, next, err := redis.HScan(client, keyForTest, cur, limit)
		if err != nil {
			t.Fatal(err)
		}
		t.Log("count:", len(results))

		for i := 0; i < len(results); i += 2 {
			field := results[i]
			value := results[i+1]
			t.Logf("key=%s, value=%s", field, value)
		}

		if next == 0 {
			break
		}
		cur = next
	}
	t.Log("hash scan done")
}

func TestReplaceHashFieldName(t *testing.T) {
	keyForTest := "test_hash"
	client, err := redis.GetRedisClientForLocalTest(t)
	if err != nil {
		t.Fatal(err)
	}

	for i := 0; i < 3; i++ {
		if err := redis.HIncrBy(client, keyForTest, "key_"+strconv.Itoa(i), int64(i)); err != nil {
			t.Fatal(err)
		}
	}

	if err := redis.ReplaceHashFieldName(client, keyForTest, "key_2", "key_2_new", 2); err != nil {
		t.Fatal(err)
	}

	results, err := redis.HGetAll(client, keyForTest)
	if err != nil {
		t.Fatal(err)
	}

	t.Log("hash count:", len(results))
	for k, v := range results {
		i, err := strconv.Atoi(v)
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("key=%s, value=%d", k, i)
	}
}

func TestSortSet(t *testing.T) {
	keyForTest := "test_sortedset"
	client, err := redis.GetRedisClientForLocalTest(t)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("zrange increase by", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			idx := rand.Intn(10)
			if err := redis.ZIncrBy(client, keyForTest, "key_"+strconv.Itoa(i), float64(idx)); err != nil {
				t.Fatal(err)
			}
		}
	})

	t.Run("zrange with scores", func(t *testing.T) {
		results, err := redis.ZRangeWithScores(client, keyForTest, 0, -1)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("count:", len(results))
		for _, item := range results {
			t.Logf("member=%s, score=%f", item.Member, item.Score)
		}
	})

}

func TestDel(t *testing.T) {
	client, err := redis.GetRedisClientForLocalTest(t)
	if err != nil {
		t.Fatal(err)
	}

	keyForTest := "test_hash"
	if err := redis.Del(client, keyForTest); err != nil {
		t.Fatal(err)
	}
	t.Log("deleted")
}
