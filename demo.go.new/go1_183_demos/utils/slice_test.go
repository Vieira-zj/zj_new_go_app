package utils_test

import (
	"testing"

	"demo.apps/utils"
)

func TestDelFirstNItemsOfSlice(t *testing.T) {
	makeSlice := func() []any {
		s := make([]any, 0, 10)
		for i := 0; i < 10; i++ {
			s = append(s, i)
		}
		return s
	}

	n := 4

	t.Run("case1", func(t *testing.T) {
		s := makeSlice()
		res, err := utils.DelFirstNItemsOfSlice(s, n)
		if err != nil {
			t.Fatal(err)
		}

		t.Log("results:", len(res), res)
	})

	t.Run("case2", func(t *testing.T) {
		s := makeSlice()
		s = s[n:]
		t.Log("results:", len(s), s)
	})
}
