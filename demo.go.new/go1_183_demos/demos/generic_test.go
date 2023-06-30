package demos

import (
	"fmt"
	"testing"
)

func TestAddByReflect(t *testing.T) {
	res, err := addByReflect(1, 1.0)
	t.Logf("result=%d, error=%s", res, err)

	res, err = addByReflect(int32(1), int32(2))
	if err != nil {
		t.Fatal(err)
	}
	t.Log("result:", res)

	res, err = addByReflect(float32(1.1), float32(2.2))
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("result: %.2f", res)
}

func TestAdd(t *testing.T) {
	res1 := add(int32(1), 1)
	t.Log("result:", res1)

	res2 := add(float32(1.0), 2.1)
	t.Log("result:", res2)

	// error
	// res3 := AddByGeneric(float32(1.0), int32(1))
	// t.Log("result:", res3)
}

func TestMin(t *testing.T) {
	t.Log("min:", min(1, 3))
	t.Log("min", min(float64(3.1), float64(2.4)))
}

type SampleSlice[T any] []T

func TestGenericSlice(t *testing.T) {
	list := make(SampleSlice[int], 2)

	list[0] = 1
	list[1] = 2
	fmt.Printf("%d\n", list[0])
	fmt.Printf("%d\n", list[1])
}

func TestSliceInsertAt(t *testing.T) {
	list1 := SampleSlice[int]{1, 2, 3, 4, 5, 6}
	res1, err := insertAt(list1, 3, 99)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res1)

	// list2 := []interface{}{}
	list2 := []any{
		"one", "two", "three", "four", "five",
	}
	res2, err := insertAt(list2, 3, "ten")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res2)
}

func TestGetFieldInfo(t *testing.T) {
	i := 9
	fmt.Println(getFieldInfo(i))

	j := &i
	fmt.Println(getFieldInfo(j))

	s := "hello"
	fmt.Println(getFieldInfo(s))

	l := []string{"foo", "bar"}
	fmt.Println(getFieldInfo(l))
}

func TestGenericKvMap(t *testing.T) {
	m := MyKvMap[string, int]{}
	m.Set("one", 1)
	m.Set("two", 2)
	m.Pprint()
}
