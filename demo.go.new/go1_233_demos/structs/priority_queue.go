package structs

import "container/heap"

type Task struct {
	ID       string
	Priority int // 数值越大, 优先级越高
	Index    int // 用于 heap 内部维护
}

// PriorityQueue 实现了 heap.Interface 接口
var _ heap.Interface = (*PriorityQueue)(nil)

type PriorityQueue []*Task

func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index, pq[j].Index = i, j
}

func (pq *PriorityQueue) Push(x any) {
	task := x.(*Task)
	task.Index = len(*pq)
	*pq = append(*pq, task)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	n := len(*pq)
	task := old[n-1]
	old[n-1] = nil  // 避免内存泄漏
	task.Index = -1 // 标记为已移除
	*pq = old[:n-1]
	return task
}
