package redis_test

import (
	"math/rand"
	"strconv"
	"testing"

	"demo.apps/middlewares/redis"
)

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
