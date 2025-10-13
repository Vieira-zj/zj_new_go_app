package structs

import "iter"

type Node[T any] struct {
	value T
	next  *Node[T]
}

type LinkedList[T any] struct {
	head *Node[T]
	tail *Node[T]
}

func (l *LinkedList[T]) Add(value T) {
	newNode := &Node[T]{value: value}
	if l.tail == nil {
		l.head = newNode
		l.tail = newNode
	} else {
		l.tail.next = newNode
		l.tail = newNode
	}
}

func (l *LinkedList[T]) Size() int {
	size := 0
	if l == nil {
		return size
	}
	for node := l.head; node != nil; node = node.next {
		size++
	}
	return size
}

// range func
// 针对遍历到的每一项都调用一次 yield 函数.
// 调用 yield 函数得到的返回值, 被用来控制循环是否继续, 若返回 true 则继续, 返回 false 则结束.

func (l *LinkedList[T]) AllValues() iter.Seq[T] {
	return func(yield func(T) bool) {
		for node := l.head; node != nil; node = node.next {
			if !yield(node.value) {
				break
			}
		}
	}
}

func (l *LinkedList[T]) AllItems() iter.Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for index, node := 0, l.head; node != nil; index, node = index+1, node.next {
			if !yield(index, node.value) {
				break
			}
		}
	}
}

// Algorithm

// SwapLinkedList swaps a LinkedList from `1 -> 2 -> 3 -> 4` to `2 -> 1 -> 4 -> 3`.
func SwapLinkedList[T any](list *LinkedList[T]) *LinkedList[T] {
	if list.Size() < 2 {
		return list
	}

	dummy := &Node[T]{next: list.head}
	curr := dummy

	for {
		first := curr.next
		second := first.next
		// swap
		first.next = second.next
		second.next = first
		curr.next = second
		// move to next pair
		curr = first
		if curr.next == nil || curr.next.next == nil {
			break
		}
	}

	return &LinkedList[T]{
		head: dummy.next,
	}
}
