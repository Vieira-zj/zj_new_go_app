package structs

import (
	"container/heap"
)

// Heap Guide:
// https://zhuanlan.zhihu.com/p/399460271
//
// Refer:
// - by use "container/heap"
// https://github.com/nsqio/nsq/blob/master/internal/pqueue/pqueue.go
// - without using "container/heap"
// https://github.com/nsqio/nsq/blob/master/nsqd/in_flight_pqueue.go
//

func DeferredQueuePush(queue heap.Interface, value any) {
	heap.Push(queue, value)
}

func DeferredQueuePop(queue heap.Interface) any {
	return heap.Pop(queue)
}

func DeferredQueueRemoveAtIndex(queue heap.Interface, idx int) {
	heap.Remove(queue, idx)
}

// PriorityQueue impls heap.Interface

type Item struct {
	Value    any
	Index    int
	Priority int
}

// PriorityQueue this is a priority queue as implemented by a min heap.
// ie. the 0th element is the *lowest* value.
type PriorityQueue []*Item

func NewPriorityQueue(capacity int) PriorityQueue {
	return make(PriorityQueue, 0, capacity)
}

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority < pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = j
	pq[j].Index = i
}

func (pq *PriorityQueue) Push(value any) {
	l := len(*pq)
	c := cap(*pq)
	if l+1 > c {
		npq := make(PriorityQueue, l, c*2)
		copy(npq, *pq)
		*pq = npq
	}

	*pq = (*pq)[:l+1]
	item := value.(*Item)
	item.Index = l
	(*pq)[l] = item
	// why not use append? *pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	l := len(*pq)
	c := cap(*pq)
	if l < (c/2) && c > 25 {
		npq := make(PriorityQueue, l, c/2)
		copy(npq, *pq)
		*pq = npq
	}

	item := (*pq)[l-1]
	item.Index = -1
	*pq = (*pq)[:l-1]
	return item
}

// PeekAndShift returns the top item whose priority is ge input max.
func (pq *PriorityQueue) PeekAndShift(max int) (*Item, int) {
	if pq.Len() == 0 {
		return nil, 0
	}

	item := (*pq)[0]
	if item.Priority > max {
		return nil, item.Priority - max
	}

	heap.Remove(pq, 0)
	return item, 0
}
