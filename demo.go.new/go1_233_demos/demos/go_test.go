package demos

import (
	"cmp"
	"encoding/json"
	"maps"
	"os"
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Common

func TestCommonSlice(t *testing.T) {
	t.Run("case1: slice append when iterator", func(t *testing.T) {
		s := []int{1, 2, 3, 4, 5}
		for _, v := range s {
			t.Log("value:", v)
			if v == 2 || v == 4 {
				s = append(s, v+10, v+20, v+30)
			}
		}
		t.Log("slice:", s)
	})

	t.Run("case2: slice append when iterator", func(t *testing.T) {
		s := []int{1, 2, 3, 4, 5}
		for len(s) > 0 {
			v := s[0]
			t.Log("value:", v)
			if v == 2 || v == 4 {
				s = append(s, v+10)
			}
			s = s[1:]
		}
		t.Log("slice:", s)
	})
}

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

// Demo: Json

func TestJsonTags(t *testing.T) {
	type Person struct {
		ID    int    `json:"id,string"`
		Name  string `json:"name"`
		Level int    `json:"level,omitzero"`
		Desc  string `json:"description,omitempty"`
	}

	t.Run("json marshal with tags", func(t *testing.T) {
		p := Person{
			ID:    102,
			Name:  "Foo",
			Level: 31,
			Desc:  "A person description",
		}
		b, err := json.Marshal(&p)
		assert.NoError(t, err)
		t.Log("json:", string(b))
	})

	t.Run("json marshal with omit tags", func(t *testing.T) {
		p := Person{
			ID:   102,
			Name: "Foo",
		}
		b, err := json.Marshal(&p)
		assert.NoError(t, err)
		t.Log("json:", string(b))
	})
}
