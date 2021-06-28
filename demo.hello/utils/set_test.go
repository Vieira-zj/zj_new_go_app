package utils

import (
	"fmt"
	"strings"
	"testing"
)

func TestAppendSlice(t *testing.T) {
	s1 := strings.Split("abc", "")
	s2 := strings.Split("xyz", "")
	s1 = append(s1, s2...)
	fmt.Println(strings.Join(s1, ","))
}

func TestPrintChar(t *testing.T) {
	input := "hello"
	for _, ch := range input {
		fmt.Printf("int=%d, char=%c\n", ch, ch)
	}
}

func TestNewSet(t *testing.T) {
	input := "this is a set test"
	set := NewSet(len(input), 'x', 'y')
	for _, word := range strings.Split(input, " ") {
		for _, ch := range word {
			set.Add(ch)
		}
	}

	fmt.Println("set count:", set.Len())
	fmt.Println("set values:")
	for _, val := range set.ToSlice() {
		fmt.Printf("%c, ", val)
	}
	fmt.Println()
}

func TestSetIntersect(t *testing.T) {
	set1, set2 := createSets()
	fmt.Println("set intersect results:")
	res := set1.Intersect(set2)
	for _, val := range res.ToSlice() {
		fmt.Printf("int=%d, ch=%c\n", val, val)
	}
}

func TestSetDiff(t *testing.T) {
	set1, set2 := createSets()
	fmt.Println("set diff results:")
	res := set1.Diff(set2)
	// res := set2.Diff(set1)
	for _, val := range res.ToSlice() {
		fmt.Printf("%c, ", val)
	}
	fmt.Println()
}

func TestSetUnion(t *testing.T) {
	set1, set2 := createSets()
	fmt.Println("set union results:")
	res := set1.Union(set2)
	for _, ch := range res.ToSlice() {
		fmt.Printf("%c, ", ch)
	}
	fmt.Println()
}

func createSets() (*Set, *Set) {
	input1 := "abcdefgabc"
	set1 := NewSet(len(input1), '1')
	for _, ch := range input1 {
		set1.Add(ch)
	}

	input2 := "xyzbcdtsxyz"
	set2 := NewSet(len(input2))
	for _, ch := range input2 {
		set2.Add(ch)
	}
	return set1, set2
}
