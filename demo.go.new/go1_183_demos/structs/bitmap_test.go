package structs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBitMap(t *testing.T) {
	offsets := []int{1, 3, 9, 10, 15}
	bitmap := NewBitMap(15)

	t.Run("bitmap set and get", func(t *testing.T) {
		for _, offset := range offsets {
			bitmap.Set(offset)
		}
		t.Log("bitmap size:", bitmap.Size())

		for _, offset := range offsets {
			ok := bitmap.Get(offset)
			assert.True(t, ok)
		}

		for _, offset := range []int{2, 8} {
			ok := bitmap.Get(offset)
			assert.False(t, ok)
		}
	})

	t.Run("bitmap reset and get", func(t *testing.T) {
		for _, offset := range []int{3, 10} {
			bitmap.Reset(offset)
		}

		for _, offset := range []int{3, 10} {
			ok := bitmap.Get(offset)
			assert.False(t, ok)
		}

		for _, offset := range []int{1, 9, 15} {
			ok := bitmap.Get(offset)
			assert.True(t, ok)
		}
	})

	t.Run("bitmap out of range", func(t *testing.T) {
		ok := bitmap.Get(71)
		assert.False(t, ok)

		err := bitmap.Set(71)
		assert.Error(t, err, "out of range")
		t.Log("error:", err)
	})
}
