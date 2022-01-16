package utils

import (
	"fmt"
	"strconv"
	"testing"
)

func TestDeleteMapValue(t *testing.T) {
	m := make(map[int]string, 3)
	for i := 0; i < 3; i++ {
		m[i] = fmt.Sprintf("%ds", i)
	}
	fmt.Println(m)

	delete(m, 1)
	fmt.Println(m)
}

func TestLruCachePut(t *testing.T) {
	cache := NewLruCache(11)
	for i := 0; i < 10; i++ {
		cache.Put(i, strconv.Itoa(i))
	}
	fmt.Println(cache.String())

	cache.Put(14, "14")
	cache.Put(17, "17")
	fmt.Println(cache.String())

	cache.Put(2, "two")
	fmt.Println(cache.String())

	cache.Put(9, "9")
	cache.Put(2, "two")
	fmt.Println(cache.String())

	cache.Clear()
	fmt.Println("size:", cache.Size())
}

func TestLruCacheGut(t *testing.T) {
	cache := NewLruCache(11)
	for i := 0; i < 10; i++ {
		cache.Put(i, strconv.Itoa(i))
	}

	for _, k := range [2]int{2, 7} {
		fmt.Printf("%d=%v\n", k, cache.Get(k))
	}
	fmt.Println(cache.String())

	cache.Put(10, "10")
	cache.Get(6)
	fmt.Println(cache.String())
}
