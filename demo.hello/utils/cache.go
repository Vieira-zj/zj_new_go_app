package utils

import (
	"fmt"
	"sync"
)

// Cache local kv cache.
type Cache struct {
	shardNumber int
	mapSize     int
	store       []map[string]interface{}
	lockers     []*sync.RWMutex
}

// NewCache creates an instance of cache.
func NewCache(shardNumber, mapSize int) *Cache {
	c := &Cache{
		shardNumber: shardNumber,
		mapSize:     mapSize,
		store:       make([]map[string]interface{}, shardNumber),
		lockers:     make([]*sync.RWMutex, shardNumber),
	}
	return c
}

// Put adds a kv in cache.
func (c *Cache) Put(key string, value interface{}) {
	locker := c.getLocker(key)
	locker.Lock()
	defer locker.Unlock()

	m := c.getMap(key)
	if _, ok := m[key]; ok {
		fmt.Printf("update existing key [%s] value\n", key)
	}
	m[key] = value
}

// Get returns value by key vaule.
func (c *Cache) Get(key string) (interface{}, error) {
	locker := c.getLocker(key)
	locker.RLock()
	defer locker.RUnlock()

	m := c.getMap(key)
	if val, ok := m[key]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("[%s] not found", key)
}

func (c *Cache) getLocker(key string) *sync.RWMutex {
	k := c.getHashKey(key)
	if c.lockers[k] == nil {
		c.lockers[k] = &sync.RWMutex{}
	}
	return c.lockers[k]
}

func (c *Cache) getMap(key string) map[string]interface{} {
	k := c.getHashKey(key)
	if c.store[k] == nil {
		c.store[k] = make(map[string]interface{}, c.mapSize)
	}
	return c.store[k]
}

func (c *Cache) getHashKey(key string) int {
	count := 0
	for _, c := range key {
		count += int(c)
	}
	return count % c.shardNumber
}

// Size returns size of cache items.
func (c *Cache) Size() int {
	size := 0
	for _, m := range c.store {
		if m != nil {
			size += len(m)
		}
	}
	return size
}

// GetItems returns all key and value pairs of cache.
func (c *Cache) GetItems() map[string]interface{} {
	for _, locker := range c.lockers {
		if locker != nil {
			locker.RLock()
		}
	}
	defer func() {
		for _, locker := range c.lockers {
			if locker != nil {
				locker.RUnlock()
			}
		}
	}()

	items := make(map[string]interface{}, c.Size())
	for _, m := range c.store {
		for k, v := range m {
			items[k] = v
		}
	}
	return items
}

// PrintKeyValues prints all keys and values of cache.
func (c *Cache) PrintKeyValues() {
	// allow data inconsistent and no lock here.
	for idx, m := range c.store {
		fmt.Printf("map%d [%d] items:\n", idx, len(m))
		for k, v := range m {
			fmt.Printf("%s=%v,", k, v)
		}
		fmt.Println()
	}
}
