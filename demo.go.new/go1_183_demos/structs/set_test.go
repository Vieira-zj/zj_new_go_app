package structs

import (
	"testing"
)

func TestSetRange(t *testing.T) {
	s := NewSet[int]()
	for n := range []int{1, 2, 3, 4, 5} {
		s.Add(n)
	}

	s.Range(func(v int) bool {
		t.Log("element:", v)
		return true
	})
}

func TestSetIterator(t *testing.T) {
	s := NewSet[int]()
	for n := range []int{1, 2, 3, 4, 5, 6} {
		s.Add(n)
	}

	next, stop := s.Iterator()
	defer stop()
	for {
		if v, ok := next(); ok {
			t.Log("element:", v)
		} else {
			break
		}
	}
}
