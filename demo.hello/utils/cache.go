package utils

import (
	"fmt"
	"strconv"
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
		store:       make([]map[string]interface{}, shardNumber),
		lockers:     make([]*sync.RWMutex, shardNumber),
		shardNumber: shardNumber,
		mapSize:     mapSize,
	}
	return c
}

// Put adds a kv in cache.
func (c *Cache) Put(key int, value interface{}) {
	locker := c.getLocker(key)
	locker.Lock()
	defer locker.Unlock()

	m := c.getMap(key)
	if _, ok := m[strconv.Itoa(key)]; ok {
		fmt.Printf("update existing key [%d] value\n", key)
	}
	m[strconv.Itoa(key)] = value
}

// Get returns value by key vaule.
func (c *Cache) Get(key int) (interface{}, error) {
	locker := c.getLocker(key)
	locker.RLock()
	defer locker.RUnlock()

	m := c.getMap(key)
	if val, ok := m[strconv.Itoa(key)]; ok {
		return val, nil
	}
	return nil, fmt.Errorf("[%d] not found", key)
}

func (c *Cache) getLocker(key int) *sync.RWMutex {
	k := c.getHashKey(key)
	if c.lockers[k] == nil {
		c.lockers[k] = &sync.RWMutex{}
	}
	return c.lockers[k]
}

func (c *Cache) getMap(key int) map[string]interface{} {
	k := c.getHashKey(key)
	if c.store[k] == nil {
		c.store[k] = make(map[string]interface{}, c.mapSize)
	}
	return c.store[k]
}

func (c *Cache) getHashKey(key int) int {
	return key % c.shardNumber
}

// Size returns size of cache.
func (c *Cache) Size() int {
	size := 0
	for _, m := range c.store {
		size += len(m)
	}
	return size
}

// GetItems returns all items of cache.
func (c *Cache) GetItems() []interface{} {
	for _, locker := range c.lockers {
		locker.RLock()
	}
	defer func() {
		for _, locker := range c.lockers {
			locker.RUnlock()
		}
	}()

	retItems := make([]interface{}, 0, c.Size())
	for _, m := range c.store {
		for _, v := range m {
			retItems = append(retItems, v)
		}
	}
	return retItems
}

// PrintItems returns all items of cache.
func (c *Cache) PrintItems() {
	// allow data inconsistent and no lock here.
	for idx, m := range c.store {
		fmt.Printf("map [%d] items:\n", idx)
		for k, v := range m {
			fmt.Printf("%s=%v,", k, v)
		}
		fmt.Println()
	}
}
