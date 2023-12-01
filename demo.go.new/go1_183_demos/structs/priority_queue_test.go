package structs_test

import (
	"math/rand"
	"sort"
	"testing"

	"demo.apps/structs"
)

func TestDeferredQueueReSize(t *testing.T) {
	c := 100
	pq := structs.NewPriorityQueue(c)
	for i := 0; i < c+1; i++ {
		item := &structs.Item{
			Value:    i,
			Priority: i,
		}
		structs.DeferredQueuePush(&pq, item)
	}
	t.Logf("len=%d, cap=%d", len(pq), cap(pq))

	for i := 0; i <= c/2+2; i++ {
		item := structs.DeferredQueuePop(&pq).(*structs.Item)
		t.Log("pop item:", item.Value)
	}
	t.Logf("len=%d, cap=%d", len(pq), cap(pq))
}

func TestDeferredQueueSort(t *testing.T) {
	c := 10
	pq := structs.NewPriorityQueue(c)
	ints := make([]int, 0, c)

	for i := 0; i < c; i++ {
		n := rand.Intn(100)
		ints = append(ints, n)
		item := &structs.Item{
			Value:    i,
			Priority: n,
		}
		structs.DeferredQueuePush(&pq, item)
	}

	sort.Ints(ints)
	t.Log("numbers:", ints)

	for i := 0; i < c; i++ {
		item := structs.DeferredQueuePop(&pq).(*structs.Item)
		t.Log(item.Priority, item.Value)
		if item.Priority != ints[i] {
			t.Fatal("unsorted deferred queue")
		}
	}
}

func TestDeferredQueuePeekAndShift(t *testing.T) {
	c := 10
	pq := structs.NewPriorityQueue(c)
	ints := make([]int, 0, c)

	for i := 0; i < c; i++ {
		n := rand.Intn(100)
		ints = append(ints, n)
		item := &structs.Item{
			Value:    i,
			Priority: n,
		}
		structs.DeferredQueuePush(&pq, item)
	}

	sort.Ints(ints)
	t.Log("numbers:", ints)

	for i := 0; i < 3; i++ {
		max := ints[0]
		if i == 2 {
			max -= 1
		}
		item, _ := pq.PeekAndShift(max)
		if item != nil {
			t.Log("peek and shift item:", item.Priority, item.Value)
			ints = ints[1:]
		}
	}

	for i := 0; i < c-2; i++ {
		item := structs.DeferredQueuePop(&pq).(*structs.Item)
		t.Log(item.Priority, item.Value)
		if item.Priority != ints[i] {
			t.Fatal("unsorted deferred queue")
		}
	}
}

func TestDeferredQueueRemoveAtIndex(t *testing.T) {
	ints := [5]int{1, 5, 2, 10, 7}
	c := len(ints)
	pq := structs.NewPriorityQueue(c)

	for i, n := range ints {
		item := &structs.Item{
			Value:    i,
			Priority: n,
		}
		structs.DeferredQueuePush(&pq, item)
	}

	idx := 1
	structs.DeferredQueueRemoveAtIndex(&pq, idx)

	for i := 0; i < len(ints)-1; i++ {
		item := structs.DeferredQueuePop(&pq).(*structs.Item)
		t.Log(item.Priority, item.Value)
	}
}
