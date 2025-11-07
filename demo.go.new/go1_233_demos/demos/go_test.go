package demos

import (
	"cmp"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"math"
	"os"
	"slices"
	"strconv"
	"time"

	"testing"
	"testing/quick"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
)

// Demo: Common Utils

type testNumbers []string

func (n *testNumbers) AppendOne(num string) {
	*n = append(*n, num) // here, need to use pointer
}

func TestCommonUtils(t *testing.T) {
	t.Run("minus uint", func(t *testing.T) {
		a, b := uint32(1), uint32(10)
		t.Log("minus uint32:", int(a-b)) // it will be overflow

		x, y := uint(1), uint(10)
		t.Log("minus uint:", int(x-y)) // it will be ok
	})

	t.Run("round float", func(t *testing.T) {
		f := 3.61
		t.Log("math floor", math.Floor(f))
		t.Log("math round", math.Round(f))

		f = 3.14169
		t.Log("round with 2 points:", strconv.FormatFloat(f, 'f', 2, 64))
		t.Log("round with 3 points:", strconv.FormatFloat(f, 'f', 3, 64))
	})

	t.Run("self type slice append", func(t *testing.T) {
		numbers := testNumbers([]string{"1", "2"})
		numbers.AppendOne("11")
		numbers.AppendOne("12")
		t.Log("numbers:", numbers)
	})

	t.Run("slice case1: append when iterator", func(t *testing.T) {
		s := []int{1, 2, 3, 4, 5}
		for _, v := range s {
			t.Log("value:", v)
			if v == 2 || v == 4 {
				s = append(s, v+10, v+20, v+30)
			}
		}
		t.Log("slice:", s)
	})

	t.Run("slice case2: append when iterator", func(t *testing.T) {
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

// Demo: Testing Mod

func TestSyncDemo(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now().UTC()
		time.Sleep(5 * time.Second) // do not block here
		t.Log("duration:", time.Since(start).Milliseconds())
	})

	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.TODO())
		afterFuncCalled := false

		context.AfterFunc(ctx, func() {
			afterFuncCalled = true
		})

		cancel()
		synctest.Wait()
		t.Logf("afterFuncCalled=%v", afterFuncCalled)
	})
}

func TestQuickDemo(t *testing.T) {
	config := quick.Config{
		MaxCount: 1000,
	}
	f := func(x, y int) bool {
		return x+y == y+x
	}

	start := time.Now()
	err := quick.Check(f, &config)
	assert.NoError(t, err)
	t.Log("done, duration:", time.Since(start).Milliseconds())
}

// Demo: Built-In Mods

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
		assert := assert.New(t)
		s := []int{1, 2, 3}
		ok := slices.Contains(s, 2)
		assert.True(ok)

		ok = slices.Contains(s, 4)
		assert.False(ok)
	})
}

func TestBuiltInIteratorOp(t *testing.T) {
	t.Run("slice iterator", func(t *testing.T) {
		slice := []int{1, 2, 3}
		it := slices.All(slice)
		for idx, val := range it {
			t.Logf("index=%d, value=%d\n", idx, val)
		}
	})

	t.Run("map iterator", func(t *testing.T) {
		s := []string{"zero", "one", "two"}
		it := slices.All(s)
		m := maps.Collect(it)
		assert.Equal(t, 3, len(m))
		t.Logf("map: %+v", m)
	})
}

func TestBuiltInSortOp(t *testing.T) {
	t.Run("sort uint32 slice", func(t *testing.T) {
		s := []uint32{21, 22, 1, 2, 3, 4}
		slices.SortFunc(s, func(a, b uint32) int {
			// return int(a - b) // it will be overflow
			return cmp.Compare(a, b)
		})
		t.Log("sorted uint32 slice:", s)
	})
}

func TestTimeUtils(t *testing.T) {
	t.Run("loop by time ticker", func(t *testing.T) {
		ctx, cancel := context.WithTimeoutCause(t.Context(), 3*time.Second, fmt.Errorf("timeout exceed"))
		defer cancel()

		tick := time.NewTicker(time.Second)
		defer tick.Stop()

		for range 10 {
			select {
			case <-ctx.Done():
				t.Logf("cancel: err=%v, cause=%v", ctx.Err(), context.Cause(ctx))
				return
			case <-tick.C:
				t.Log("after 1 second")
			}
		}
		t.Log("done")
	})
}

func TestOsUtils(t *testing.T) {
	t.Run("os exec", func(t *testing.T) {
		path, err := os.Executable()
		assert.NoError(t, err)
		t.Log("exec path:", path)
	})
}

// Demo: Json

func TestJsonMarshalTags(t *testing.T) {
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
