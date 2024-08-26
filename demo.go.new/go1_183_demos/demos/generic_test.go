package demos

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddByType(t *testing.T) {
	t.Run("add by type, fail with error", func(t *testing.T) {
		res, err := addT(1, 2)
		assert.Error(t, err)
		t.Logf("result=%d, error=%s", res, err)
	})

	t.Run("add by type, pass with int32", func(t *testing.T) {
		res, err := addT(int32(1), int32(2))
		assert.NoError(t, err)
		t.Log("result:", res)
	})
}

func TestAddByReflect(t *testing.T) {
	t.Run("add by reflect, fail with error", func(t *testing.T) {
		res, err := addR(1, 1.0)
		assert.Error(t, err)
		t.Logf("result=%d, error=%s", res, err)
	})

	t.Run("add by reflect, pass with int32", func(t *testing.T) {
		res, err := addR(int32(1), int32(2))
		assert.NoError(t, err)
		t.Log("result:", res)
	})

	t.Run("add by reflect, pass with float32", func(t *testing.T) {
		res, err := addR(float32(1.1), float32(2.2))
		assert.NoError(t, err)
		t.Logf("result: %.2f", res)
	})
}

func TestAddByGeneric(t *testing.T) {
	//nolint:unused
	type testInt32 int

	res1 := addG(int32(1), 1)
	t.Log("result:", res1)

	res2 := addG(float32(1.0), 2.1)
	t.Log("result:", res2)

	// res3 := addG(TestInt32(1), TestInt32(1))
	// t.Log("result:", res3)

	// res3 := addG(float32(1.0), int32(1))
	// t.Log("result:", res3)
}

func TestGenericMin(t *testing.T) {
	type testInt int

	t.Log("min:", min[int](1, 3))
	t.Log("min:", min(testInt(1), testInt(3))) // it's ok because of ~int
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

func TestGenericSliceInsertAt(t *testing.T) {
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

func TestGenericGetFieldInfo(t *testing.T) {
	i := 9
	fmt.Println(getFieldInfo(i))

	j := &i
	fmt.Println(getFieldInfo(j))

	s := "hello"
	fmt.Println(getFieldInfo(s))

	l := []string{"foo", "bar"}
	fmt.Println(getFieldInfo(l))
}

func TestGenericPlusScalar(t *testing.T) {
	type myStr string

	t.Log("1+2:", plus(1, 2))
	t.Log("1.5 + 2.7:", plus(1.5, 2.7))
	t.Log("string cat:", plus("hello, ", "world"))
	t.Log("my string cat:", plus(myStr("Go, "), myStr("1.18")))
}

func TestGenericKvMap(t *testing.T) {
	m := MyKvMap[string, int]{}
	m.Set("one", 1)
	m.Set("two", 2)
	m.Pprint()
}
