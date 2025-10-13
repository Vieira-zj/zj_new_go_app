package structs

import "testing"

func TestLinkedList(t *testing.T) {
	t.Run("linked list values iterator", func(t *testing.T) {
		l := &LinkedList[int]{}
		for i := 3; i <= 10; i++ {
			l.Add(i)
		}
		for v := range l.AllValues() {
			t.Logf("value: %d", v)
		}
	})

	t.Run("linked list key-values iterator", func(t *testing.T) {
		l := &LinkedList[int]{}
		for i := range 10 {
			l.Add(i)
		}

		for i, v := range l.AllItems() {
			t.Logf("index: %d, value: %d", i, v)
		}
	})
}

func TestSwapLinkedList(t *testing.T) {
	l := &LinkedList[int]{}
	for i := 1; i <= 4; i++ {
		l.Add(i)
	}

	t.Log("before swap:")
	for v := range l.AllValues() {
		t.Logf("value: %d", v)
	}

	result := SwapLinkedList(l)

	t.Log("after swap:")
	for v := range result.AllValues() {
		t.Logf("value: %d", v)
	}
}
