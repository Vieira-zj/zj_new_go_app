package utils

import (
	"fmt"
	"strconv"
	"sync"
)

/*
ShardMap
*/

// Cache local kv cache.
type Cache struct {
	shard   int
	mapSize int
	store   []map[string]interface{}
	lockers []*sync.RWMutex
}

// NewCache creates an instance of cache.
func NewCache(shardNumber, mapSize int) *Cache {
	c := &Cache{
		shard:   shardNumber,
		mapSize: mapSize,
		store:   make([]map[string]interface{}, shardNumber),
		lockers: make([]*sync.RWMutex, shardNumber),
	}
	return c
}

// Put adds a kv in cache. if key exist, then overwrite.
func (c *Cache) Put(key string, value interface{}) {
	locker := c.getLocker(key)
	locker.Lock()
	defer locker.Unlock()

	m := c.getMap(key)
	if _, ok := m[key]; ok {
		fmt.Printf("Update existing key [%s] value\n", key)
	}
	m[key] = value
}

// PutIfEmpty adds a kv in cache when key is not exist.
func (c *Cache) PutIfEmpty(key string, value interface{}) {
	locker := c.getLocker(key)
	locker.Lock()
	defer locker.Unlock()

	m := c.getMap(key)
	if _, ok := m[key]; ok {
		return
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
	return nil, fmt.Errorf("Key [%s] not found", key)
}

// IsExist returns whether key is exist.
func (c *Cache) IsExist(key string) bool {
	locker := c.getLocker(key)
	locker.RLock()
	defer locker.RUnlock()

	m := c.getMap(key)
	_, ok := m[key]
	return ok
}

func (c *Cache) getLocker(key string) *sync.RWMutex {
	k := c.getHashKey(key)
	if c.lockers[k] == nil {
		c.lockers[k] = new(sync.RWMutex)
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
	// TODO: add salt by random number or nano time.
	count := 0
	for _, c := range key {
		count += int(c)
	}
	return count % c.shard
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

// UsageToText .
func (c *Cache) UsageToText() string {
	line := ""
	total := 0
	for i, m := range c.store {
		total += len(m)
		line += fmt.Sprintf("map%d:%d/%d | ", i, len(m), c.mapSize)
	}
	line += fmt.Sprintln("total:" + strconv.Itoa(total))
	return line
}
