package structs

import (
	"hash/maphash"
	"sync"
)

// RWMutex Sync Map

type MutexSyncMap[T any] struct {
	mux  sync.RWMutex
	data map[string]T
}

func NewMutexSyncMap[T any]() MutexSyncMap[T] {
	return MutexSyncMap[T]{
		data: make(map[string]T),
	}
}

func (m *MutexSyncMap[T]) Set(key string, v T) {
	m.mux.Lock()
	defer m.mux.Unlock()
	m.data[key] = v
}

func (m *MutexSyncMap[T]) Get(key string) (T, bool) {
	m.mux.RLock()
	defer m.mux.RUnlock()
	v, ok := m.data[key]
	return v, ok
}

func (m *MutexSyncMap[T]) Del(key string) {
	m.mux.Lock()
	defer m.mux.Unlock()
	delete(m.data, key)
}

// Sharding Sync Map

type ShardingSyncMap[T any] struct {
	shards []MutexSyncMap[T]
}

func NewShardingSyncMap[T any](size int) ShardingSyncMap[T] {
	shards := make([]MutexSyncMap[T], size)
	for i := 0; i < size; i++ {
		shards[i] = MutexSyncMap[T]{
			data: make(map[string]T),
		}
	}
	return ShardingSyncMap[T]{
		shards: shards,
	}
}

func (m *ShardingSyncMap[T]) Set(key string, value T) {
	idx := m.getShardIdx(key)
	m.shards[idx].Set(key, value)
}

func (m *ShardingSyncMap[T]) Get(key string) (T, bool) {
	idx := m.getShardIdx(key)
	return m.shards[idx].Get(key)
}

func (m *ShardingSyncMap[T]) Del(key string) {
	idx := m.getShardIdx(key)
	m.shards[idx].Del(key)
}

var seed = maphash.MakeSeed()

func (m *ShardingSyncMap[T]) getShardIdx(key string) uint64 {
	hash := m.hashKey(key)
	return hash % uint64(len(m.shards))
}

func (ShardingSyncMap[T]) hashKey(key string) uint64 {
	return maphash.String(seed, key)
}

// Channel Sync Map

type ChanSyncMap[T any] struct {
	cmd  chan command[T]
	data map[string]T
}

type command[T any] struct {
	action string // "get", "set", "delete"
	key    string
	value  T
	result chan<- result[T]
}

type result[T any] struct {
	value T
	ok    bool
}

func NewChanSyncMap[T any]() *ChanSyncMap[T] {
	chanMap := &ChanSyncMap[T]{
		cmd:  make(chan command[T]),
		data: make(map[string]T),
	}

	go chanMap.run()
	return chanMap
}

func (m *ChanSyncMap[T]) run() {
	for cmd := range m.cmd {
		switch cmd.action {
		case "get":
			value, ok := m.data[cmd.key]
			cmd.result <- result[T]{value, ok}
		case "set":
			m.data[cmd.key] = cmd.value
		case "delete":
			delete(m.data, cmd.key)
		}
	}
}

func (m *ChanSyncMap[T]) Set(key string, value T) {
	m.cmd <- command[T]{action: "set", key: key, value: value}
}

func (m *ChanSyncMap[T]) Get(key string) (T, bool) {
	res := make(chan result[T])
	m.cmd <- command[T]{action: "get", key: key, result: res}
	r := <-res
	return r.value, r.ok
}

func (m *ChanSyncMap[T]) Del(key string) {
	m.cmd <- command[T]{action: "delete", key: key}
}
