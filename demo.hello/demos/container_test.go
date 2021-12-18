package demos

import (
	"container/heap"
	"container/list"
	"container/ring"
	"fmt"
	"strings"
	"testing"
)

/* List 双向链表 */

func printList(l *list.List) {
	for e := l.Front(); e != nil; e = e.Next() {
		fmt.Printf("%v,", e.Value)
	}
	fmt.Println()
}

func TestList(t *testing.T) {
	l := list.New()
	for _, val := range strings.Split("hellofoo", "") {
		l.PushBack(val)
	}
	printList(l)

	fmt.Println("=========")
	fmt.Println("first:", l.Front().Value)
	fmt.Println("last:", l.Back().Value)

	fmt.Println("=========")
	l.PushFront("1")
	printList(l)

	fmt.Println("=========")
	for e := l.Front(); e != nil; e = e.Next() {
		if e.Value == "1" {
			l.InsertAfter("2", e)
		}
		if e.Value == "f" {
			l.InsertBefore("_", e)
		}
	}
	printList(l)

	fmt.Println("=========")
	for e := l.Back(); e != nil; e = e.Prev() {
		fmt.Printf("%v,", e.Value)
	}
	fmt.Println()
}

/* Ring */

func TestRing(t *testing.T) {
	ring1 := ring.New(3)
	for i := 1; i <= 3; i++ {
		ring1.Value = i
		ring1 = ring1.Next()
	}

	ring2 := ring.New(3)
	for i := 4; i <= 6; i++ {
		ring2.Value = i
		ring2 = ring2.Next()
	}

	r := ring1.Link(ring2)
	fmt.Println("ring size:", r.Len())

	fmt.Println("=========")
	r.Do(func(p interface{}) {
		fmt.Printf("%d,", p.(int))
	})
	fmt.Println()

	fmt.Println("=========")
	fmt.Println("current value:", r.Value)
	fmt.Println("next value:", r.Next().Value)
	fmt.Println("pre value:", r.Prev().Value)

	fmt.Println("=========")
	for p := r.Next(); p != r; p = p.Next() {
		fmt.Printf("%v,", p.Value)
	}
	fmt.Println()
}

/*
Heap

最小堆，是一种经过排序的完全二叉树，其中任一非终端节点的数据值均不大于其左子节点和右子节点的值。

- 数组来实现二叉树，所以满足二叉树的特性
- 根元素是最小的元素，父节点小于它的两个子节点
- 树中的元素是相对有序的
*/

type StudentHeapNode struct {
	name  string
	score int
}

type StudentHeap []StudentHeapNode

func (h StudentHeap) Len() int {
	return len(h)
}

func (h StudentHeap) Less(i, j int) bool {
	return h[i].score < h[j].score // 最小堆
}

func (h StudentHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *StudentHeap) Push(x interface{}) {
	// Push and Pop use pointer receivers because they modify the slice's length, not just its contents.
	*h = append(*h, x.(StudentHeapNode))
}

func (h *StudentHeap) Pop() interface{} {
	// get index value: (*h)[0], not use: *h[0]
	local := *h
	// 返回头节点，即最小值
	ret := local[0]
	*h = local[1:]
	return ret
}

func TestHeap(t *testing.T) {
	h := &StudentHeap{
		{name: "xiaoming", score: 82},
		{name: "xiaozhang", score: 88},
		{name: "laowang", score: 85},
	}

	// Init函数对于堆的约束性是幂等的（多次执行无意义），并可能在任何时候堆的约束性被破坏时被调用
	heap.Init(h)
	// 向堆h中插入元素x, 并保持堆的约束性
	heap.Push(h, StudentHeapNode{name: "xiaoli", score: 66})

	for _, ele := range *h {
		fmt.Printf("student name %s,score %d\n", ele.name, ele.score)
	}

	fmt.Println("=========")
	for i, ele := range *h {
		if ele.name == "xiaozhang" {
			(*h)[i].score = 60
			// 在修改第i个元素后，调用本函数修复堆，比删除第i个元素后插入新元素更有效率
			heap.Fix(h, i)
		}
	}
	for _, ele := range *h {
		fmt.Printf("student name %s,score %d\n", ele.name, ele.score)
	}

	fmt.Println("=========")
	for h.Len() > 0 {
		// 删除并返回堆h中的最小元素（取决于Less函数，最大堆或最小堆）（不影响堆de约束性）
		item := h.Pop().(StudentHeapNode)
		fmt.Printf("student name %s,score %d\n", item.name, item.score)
	}
}

func TestSlice(t *testing.T) {
	s := []string{"a", "b", "c"}
	first := s[len(s)-1]
	s = s[0 : len(s)-1]
	fmt.Println(first)
	fmt.Println(s)
}
