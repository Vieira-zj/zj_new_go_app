package demos

import (
	"cmp"
	"maps"
	"os"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Demo: Built-In Fn

func TestBuiltInCmpOp(t *testing.T) {
	t.Run("cmp or", func(t *testing.T) {
		result := cmp.Or(os.Getenv("SOME_VARIABLE"), "default")
		t.Log("env:", result)
	})
}

func TestBuiltInSlicesOp(t *testing.T) {
	t.Run("slices concat", func(t *testing.T) {
		s := slices.Concat([]int{1, 2}, []int{3}, []int{7, 8, 9})
		t.Log("concat slice:", s)
	})

	t.Run("slices contains", func(t *testing.T) {
		s := []int{1, 2, 3}
		ok := slices.Contains(s, 2)
		assert.True(t, ok)

		ok = slices.Contains(s, 4)
		assert.False(t, ok)
	})
}

func TestBuiltInIteratorOp(t *testing.T) {
	t.Run("iterator loop", func(t *testing.T) {
		slice := []int{1, 2, 3}
		it := slices.All(slice)
		for idx, val := range it {
			t.Logf("index=%d, value=%d\n", idx, val)
		}
	})
}

func TestBuiltInMapsOp(t *testing.T) {
	t.Run("maps collect", func(t *testing.T) {
		s := []string{"zero", "one", "two"}
		it := slices.All(s)
		m := maps.Collect(it)
		assert.Equal(t, 3, len(m))
		t.Logf("map: %+v", m)
	})
}

// Demo: Built-In Libs

func TestOsUtils(t *testing.T) {
	t.Run("os exec", func(t *testing.T) {
		path, err := os.Executable()
		assert.NoError(t, err)
		t.Log("exec path:", path)
	})
}
