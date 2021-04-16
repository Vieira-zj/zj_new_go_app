package utils

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestCachePutGet(t *testing.T) {
	shardNumber := 3
	cache := NewCache(shardNumber, 20)

	wait := sync.WaitGroup{}
	for i := 0; i < shardNumber; i++ {
		wait.Add(1)
		go func(base int) {
			defer wait.Done()
			base *= 10
			for j := base; j < base+10; j++ {
				cache.Put(j, fmt.Sprintf("value%d", j))
				time.Sleep(time.Duration(100) * time.Millisecond)
			}
		}(i)
	}
	wait.Wait()
	fmt.Printf("Cache size: %d\n", cache.Size())
	cache.PrintItems()

	for i := 0; i <= 30; i++ {
		val, err := cache.Get(i)
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Printf("get: %d=%v\n", i, val)
	}
}
