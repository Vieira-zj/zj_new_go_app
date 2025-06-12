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

		for i, v := range l.All() {
			t.Logf("index: %d, value: %d", i, v)
		}
	})
}
