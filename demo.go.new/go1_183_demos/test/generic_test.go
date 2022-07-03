package test

import (
	"fmt"
	"testing"
)

type SampleSlice[T any] []T

func TestGenericSlice(t *testing.T) {
	var list SampleSlice[int]
	list = make(SampleSlice[int], 2)

	list[0] = 1
	list[1] = 2
	fmt.Printf("%d\n", list[0])
	fmt.Printf("%d\n", list[1])
}

type SampleMap[K comparable, V any] map[K]V

func TestGenericMap(t *testing.T) {
	dict := make(SampleMap[string, int], 2)

	dict["one"] = 1
	dict["two"] = 2
	fmt.Printf("dict: %+v\n", dict)
}

func TestSliceInsertAt(t *testing.T) {
	list1 := SampleSlice[int]{1, 2, 3, 4, 5, 6}
	res1, err := InsertAt(list1, 3, 99)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res1)

	// list2 := []interface{}{}
	list2 := []any{
		"one", "two", "three", "four", "five",
	}
	res2, err := InsertAt(list2, 3, "ten")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res2)
}

func TestGetFieldInfo(t *testing.T) {
	i := 9
	fmt.Println(GetFieldInfo(i))

	j := &i
	fmt.Println(GetFieldInfo(j))

	s := "hello"
	fmt.Println(GetFieldInfo(s))

	l := []string{"foo", "bar"}
	fmt.Println(GetFieldInfo(l))
}
