package utils

import (
	"container/list"
	"fmt"
	"strings"
	"sync"
)

//
// 使用 list + map 实现，get/put 操作仍然需要遍历 list.
// 优化：使用 map + node{prev,next} 实现一个 LinkedHashMap, 避免遍历 list.
//

// LruCache Least Recently Used kv cache.
type LruCache struct {
	size         uint32
	linkedValues *list.List
	hashValues   map[interface{}]*list.Element
	locker       sync.RWMutex
}

// NewLruCache creates an instance of LruCache.
func NewLruCache(size uint32) *LruCache {
	return &LruCache{
		size:         size,
		linkedValues: list.New(),
		hashValues:   make(map[interface{}]*list.Element, size),
	}
}

// Put .
func (c *LruCache) Put(k, v interface{}) {
	c.locker.Lock()
	defer c.locker.Unlock()

	if ele, ok := c.hashValues[k]; ok {
		c.linkedValues.Remove(ele)
	}
	if c.linkedValues.Len() == int(c.size) {
		ele := c.linkedValues.Back()
		c.linkedValues.Remove(ele)
		delete(c.hashValues, k)
	}
	ele := c.linkedValues.PushFront(v)
	c.hashValues[k] = ele
}

// Get .
func (c *LruCache) Get(k interface{}) interface{} {
	c.locker.RLock()
	defer c.locker.RUnlock()

	ele, ok := c.hashValues[k]
	if !ok {
		return nil
	}
	c.linkedValues.MoveToFront(ele)
	return ele.Value
}

// Size .
func (c *LruCache) Size() uint32 {
	return uint32(c.linkedValues.Len())
}

// String .
func (c *LruCache) String() string {
	if c.Size() == 0 {
		return ""
	}

	eles := make([]string, 0, c.size)
	for _, ele := range c.AsList() {
		eles = append(eles, fmt.Sprintf("%v", ele))
	}
	return strings.Join(eles, ",")
}

// AsList .
func (c *LruCache) AsList() []interface{} {
	if c.Size() == 0 {
		return []interface{}{}
	}

	retList := make([]interface{}, 0, c.linkedValues.Len())
	for e := c.linkedValues.Front(); e != nil; e = e.Next() {
		retList = append(retList, e.Value)
	}
	return retList
}

// Clear .
func (c *LruCache) Clear() {
	c.locker.Lock()
	defer c.locker.Unlock()

	if c.Size() == 0 {
		return
	}
	c.linkedValues = list.New()
	c.hashValues = make(map[interface{}]*list.Element, c.size)
}
