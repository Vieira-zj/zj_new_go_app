package demos

import (
	"strconv"
	"testing"

	"github.com/samber/lo"
)

func TestLoSliceUnique(t *testing.T) {
	names := lo.Uniq[string]([]string{"foo", "bar", "foo"})
	t.Log("names:", names)
}

func TestLoSliceFilterAndMap(t *testing.T) {
	even := lo.Filter[int]([]int{1, 2, 3, 4}, func(x int, index int) bool {
		return x%2 == 0
	})
	t.Log("even num:", even)

	result := lo.Map[int64, string]([]int64{1, 2, 3, 4}, func(x int64, index int) string {
		return strconv.FormatInt(x, 10)
	})
	t.Log("map result:", result)
}

func TestLoMapKeysAndValues(t *testing.T) {
	m := map[string]int{"foo": 1, "bar": 2}
	keys := lo.Keys[string, int](m)
	t.Log("map keys:", keys)

	values := lo.Values[string, int](m)
	t.Log("map values:", values)

	value := lo.ValueOr[string, int](m, "test", 29)
	t.Log("value:", value)
}

func TestLoStringTest(t *testing.T) {
	sub := lo.Substring("hello", 2, 3)
	t.Log("sub str:", sub)

	str := lo.RandomString(12, lo.LettersCharset)
	t.Log("rand str:", str)
}
