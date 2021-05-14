package utils

import (
	"fmt"
	"math"
	"strconv"
	"testing"
)

func TestMath(t *testing.T) {
	fmt.Printf("inf: %d\n", math.MaxInt32)
}

func TestSkiplistNode(t *testing.T) {
	node1 := newSkiplistNode(1, "1", 2)
	fmt.Println("node:", node1.string())

	node2 := newSkiplistNode(2, "2", 1)
	if err := node1.setNextNode(node2, 3); err != nil {
		fmt.Println(err.Error())
	}
}

func TestSkiplistRandomDepth(t *testing.T) {
	s := NewSkiplist(5, 0.5)
	for i := 0; i < 6; i++ {
		fmt.Println("random depth:", s.getRandomDepth())
	}
}

func TestSkiplistInsert(t *testing.T) {
	s := NewSkiplist(5, 0.5)
	s.Insert(2, "2")
	s.Insert(1, "1")
	s.Insert(3, "3")
	s.Print()

	fmt.Println("\nupdate:")
	s.Insert(3, "3rd")
	s.Print()
}

func TestSkiplistDelete(t *testing.T) {
	s := NewSkiplist(5, 0.5)
	s.Insert(2, "2")
	s.Insert(1, "1")
	s.Insert(3, "3")
	s.Print()

	fmt.Println("\ndelete:")
	s.Delete(2)
	s.Print()
}

func TestSkiplistQuery(t *testing.T) {
	s := NewSkiplist(5, 0.5)
	for _, val := range []int{38, 55, 12, 31, 17, 50, 25, 44, 20, 39} {
		s.Insert(val, strconv.Itoa(val))
	}
	s.Print()
	res, err := s.Query(39)
	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("query result:", res)
}
