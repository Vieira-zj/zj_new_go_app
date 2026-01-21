package structs

import (
	"strconv"
	"testing"
)

func TestLRUCache(t *testing.T) {
	cache := NewLRUCache(3)
	for i := range 3 {
		cache.Put(strconv.Itoa(i), i)
	}
	for e := range cache.Iter() {
		t.Logf("key=%s, value=%v", e.key, e.value)
	}

	cache.Put("4", 4)
	cache.Put("2", 2)
	t.Log("add more element:")
	for e := range cache.Iter() {
		t.Logf("key=%s, value=%v", e.key, e.value)
	}
}
