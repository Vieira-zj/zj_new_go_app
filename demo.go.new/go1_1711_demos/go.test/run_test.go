package gotest

import (
	"strconv"
	"sync"
	"testing"
)

// go test -race 数据竞争检查

type myAccount struct {
	value   int
	records []string
	locker  sync.Mutex
}

func newMyAccount() *myAccount {
	return &myAccount{
		records: make([]string, 0, 8),
		locker:  sync.Mutex{},
	}
}

func (a *myAccount) Add(val int) {
	a.locker.Lock()
	a.value += val
	a.locker.Unlock()
}

func (a *myAccount) Min(val int) {
	a.value -= val
}

func (a *myAccount) BatchAdd(vals []int) {
	for _, val := range vals {
		val := val
		go func() {
			a.Add(val)
		}()
	}
}

func (a *myAccount) BatchMin(vals []int) {
	for _, val := range vals {
		val := val
		go func() {
			a.Min(val)
		}()
	}
}

func TestBatchOp(t *testing.T) {
	vals := make([]int, 100)
	for i := 0; i < 30; i++ {
		vals[i] = 1
	}

	account := newMyAccount()
	account.BatchAdd(vals)
	account.BatchMin(vals[:10])
	t.Log("value:", account.value)
}

// go test -fuzz

func (s *myAccount) AddRecord(record string) {
	s.records = append(s.records, record)
}

func (s *myAccount) GetRecord(idx int) string {
	// if idx >= len(s.records) || idx < 0 {
	if idx > len(s.records) {
		return ""
	}
	return s.records[idx]
}

func FuzzGetRecord(f *testing.F) {
	account := newMyAccount()
	for i := 0; i < 10; i++ {
		account.AddRecord(strconv.Itoa(i))
	}

	f.Fuzz(func(t *testing.T, val int) {
		account.GetRecord(val)
	})
}
