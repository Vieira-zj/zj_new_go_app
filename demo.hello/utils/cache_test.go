package utils

import (
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestCachePut(t *testing.T) {
	shardNumber := 3
	cache := NewCache(shardNumber, 10)
	fmt.Println(cache.GetItems())

	cache.Put("1", "value1")
	cache.Put("2", "value2")
	fmt.Println("store:", cache.store)
	fmt.Println("lockers:", cache.lockers)

	fmt.Println("\nusage:")
	fmt.Println(cache.UsageToText())

	fmt.Println("\ncache values:")
	for k, v := range cache.GetItems() {
		fmt.Printf("%s=%v\n", k, v)
	}
}

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
				cache.Put(strconv.Itoa(j), fmt.Sprintf("value%d", j))
				time.Sleep(time.Duration(100) * time.Millisecond)
			}
		}(i)
	}
	wait.Wait()
	fmt.Printf("Cache size: %d\n", cache.Size())
	cache.PrintKeyValues()

	for i := 0; i <= 30; i++ {
		val, err := cache.Get(strconv.Itoa(i))
		if err != nil {
			fmt.Println(err.Error())
			continue
		}
		fmt.Printf("get: %d=%v\n", i, val)
	}
}
