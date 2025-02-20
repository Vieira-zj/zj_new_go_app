package structs

import (
	"fmt"
	"sync"
	"testing"
)

func BenchmarkSyncMap01(b *testing.B) {
	b.Log("sync map benchmark: read and write are the same")

	b.Run("go sync map", func(b *testing.B) {
		var m sync.Map
		var wg sync.WaitGroup

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(k int) {
				defer wg.Done()
				key := fmt.Sprintf("key%d", k)
				m.Store(key, k)
				m.Load(key)
			}(i % 100000)
		}
		wg.Wait()
	})

	b.Run("mutex sync map", func(b *testing.B) {
		m := NewMutexSyncMap[int]()
		var wg sync.WaitGroup

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(k int) {
				defer wg.Done()
				key := fmt.Sprintf("key%d", k)
				m.Set(key, k)
				m.Get(key)
			}(i % 100000)
		}
		wg.Wait()
	})

	b.Run("sharding sync map", func(b *testing.B) {
		m := NewShardingSyncMap[int](8)
		var wg sync.WaitGroup

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(k int) {
				defer wg.Done()
				key := fmt.Sprintf("key%d", k)
				m.Set(key, k)
				m.Get(key)
			}(i % 100000)
		}
		wg.Wait()
	})

	b.Run("chan sync map", func(b *testing.B) {
		m := NewChanSyncMap[int]()
		var wg sync.WaitGroup

		for i := 0; i < b.N; i++ {
			wg.Add(1)
			go func(k int) {
				defer wg.Done()
				key := fmt.Sprintf("key%d", k)
				m.Set(key, k)
				m.Get(key)
			}(i % 100000)
		}
		wg.Wait()
	})
}

func BenchmarkSyncMap02(b *testing.B) {
	b.Log("sync map benchmark: read more and write less")

	b.Run("go sync map", func(b *testing.B) {
		var m sync.Map
		var wg sync.WaitGroup

		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				wg.Add(1)
				go func(k int) {
					defer wg.Done()
					m.Load(string(rune(k)))
				}(j)
			}
			for j := 0; j < 10; j++ {
				wg.Add(1)
				go func(k int) {
					defer wg.Done()
					m.Store(string(rune(k)), k)
				}(j)
			}
		}
		wg.Wait()
	})

	b.Run("mutex sync map", func(b *testing.B) {
		var m = NewMutexSyncMap[int]()
		var wg sync.WaitGroup

		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				wg.Add(1)
				go func(k int) {
					defer wg.Done()
					m.Get(string(rune(k)))
				}(j)
			}
			for j := 0; j < 10; j++ {
				wg.Add(1)
				go func(k int) {
					defer wg.Done()
					m.Set(string(rune(k)), k)
				}(j)
			}
		}
		wg.Wait()
	})

	b.Run("sharding sync map", func(b *testing.B) {
		var m = NewShardingSyncMap[int](8)
		var wg sync.WaitGroup

		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				wg.Add(1)
				go func(k int) {
					defer wg.Done()
					m.Get(string(rune(k)))
				}(j)
			}
			for j := 0; j < 10; j++ {
				wg.Add(1)
				go func(k int) {
					defer wg.Done()
					m.Set(string(rune(k)), k)
				}(j)
			}
		}
		wg.Wait()
	})

	b.Run("chan sync map", func(b *testing.B) {
		var m = NewChanSyncMap[int]()
		var wg sync.WaitGroup

		for i := 0; i < b.N; i++ {
			for j := 0; j < 100; j++ {
				wg.Add(1)
				go func(k int) {
					defer wg.Done()
					m.Get(string(rune(k)))
				}(j)
			}
			for j := 0; j < 10; j++ {
				wg.Add(1)
				go func(k int) {
					defer wg.Done()
					m.Set(string(rune(k)), k)
				}(j)
			}
		}
		wg.Wait()
	})
}
