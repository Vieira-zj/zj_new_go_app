package structs

import (
	"container/list"
	"iter"
	"sync"
)

type Entry struct {
	key   string
	value any
}

// LRU: Least Recently Used
type LRUCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.Mutex
}

func NewLRUCache(capacity int) *LRUCache {
	if capacity <= 0 {
		capacity = 1
	}

	return &LRUCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element, 16),
		list:     list.New(),
	}
}

func (l *LRUCache) Size() int {
	return len(l.cache)
}

func (l *LRUCache) Get(key string) (any, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if elem, ok := l.cache[key]; ok {
		l.list.MoveToFront(elem)
		return elem.Value.(*Entry).value, true
	}
	return nil, false
}

func (l *LRUCache) Put(key string, value any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 如果 key 已存在, 更新值并移到头部
	if elem, ok := l.cache[key]; ok {
		elem.Value.(*Entry).value = value
		l.list.MoveToFront(elem)
		return
	}

	// 判断是否需要淘汰
	if l.list.Len() >= l.capacity {
		l.removeOldest()
	}

	elem := l.list.PushFront(&Entry{key: key, value: value})
	l.cache[key] = elem
}

func (l *LRUCache) Iter() iter.Seq[*Entry] {
	return func(yield func(*Entry) bool) {
		for e := l.list.Front(); e != nil; e = e.Next() {
			if !yield(e.Value.(*Entry)) {
				break
			}
		}
	}
}

func (l *LRUCache) removeOldest() {
	elem := l.list.Back()
	if elem == nil {
		return
	}

	l.list.Remove(elem)
	key := elem.Value.(*Entry).key
	delete(l.cache, key)
}
