package demos

import (
	"cmp"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"math"
	"os"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"

	"testing"
	"testing/quick"
	"testing/synctest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	t.Run("clear slice and map", func(t *testing.T) {
		s := []int{1, 2, 3, 4, 5}
		clear(s)
		assert.Len(t, s, 5)
		t.Log("slice after clear:", s)

		m := map[string]int{"one": 1, "two": 2, "three": 3}
		clear(m)
		assert.Len(t, m, 0)
		t.Log("map after clear:", m)
	})
}

// Demo: Testing Mod

func TestSyncDemo(t *testing.T) {
	t.Run("run goroutine in synctest", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			for i := range 3 {
				go func(idx int) {
					t.Logf("hello from goroutine [%d]", idx)
				}(i)
			}
			synctest.Wait()
		})
	})

	t.Run("sleep in synctest", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			start := time.Now().UTC()
			time.Sleep(5 * time.Second) // do not block here
			t.Log("duration:", time.Since(start).Milliseconds())
		})
	})

	t.Run("call ctx after_func in synctest", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.TODO())
			afterFuncCalled := false

			context.AfterFunc(ctx, func() {
				afterFuncCalled = true
			})

			go func() {
				cancel()
			}()
			synctest.Wait()
			t.Logf("is after func called=%v", afterFuncCalled)
		})
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

func TestCmpUtil(t *testing.T) {
	t.Run("cmp or", func(t *testing.T) {
		// 返回第一个非空字符串
		result := cmp.Or(os.Getenv("SOME_VARIABLE"), "default")
		t.Log("env:", result)
	})
}

func TestSlicesUtil(t *testing.T) {
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

	t.Run("sort uint32 slice", func(t *testing.T) {
		s := []uint32{21, 22, 1, 2, 3, 4}
		slices.SortFunc(s, func(a, b uint32) int {
			// return int(a - b) // it will be overflow
			return cmp.Compare(a, b)
		})
		t.Log("sorted uint32 slice:", s)
	})
}

func TestIteratorOp(t *testing.T) {
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

func TestStringsUtil(t *testing.T) {
	t.Run("string cut", func(t *testing.T) {
		s := "hello||world"
		before, after, found := strings.Cut(s, "||")
		if found {
			t.Logf("before=%s, after=%s", before, after)
		} else {
			t.Log("delimiter not found")
		}
	})
}

func TestTimeUtil(t *testing.T) {
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

func TestOsUtil(t *testing.T) {
	t.Run("os exec", func(t *testing.T) {
		path, err := os.Executable()
		assert.NoError(t, err)
		t.Log("exec path:", path)
	})

	t.Run("os open root", func(t *testing.T) {
		root, err := os.OpenRoot("/tmp/test")
		require.NoError(t, err)

		// ops base on root dir
		b, err := root.ReadFile("output.json")
		assert.NoError(t, err)
		t.Log("read file:\n", string(b))
	})
}

func TestErrorsUtil(t *testing.T) {
	t.Run("error is check", func(t *testing.T) {
		customErr := fmt.Errorf("custom test error")
		process := func(hasErr bool) error {
			if hasErr {
				return customErr
			}
			return nil
		}

		if err := process(true); err != nil {
			if errors.Is(err, customErr) {
				t.Log("get custom error")
			} else {
				t.Log(err)
			}
		}
	})

	t.Run("error type check", func(t *testing.T) {
		if _, err := os.Open("non_existing_file.txt"); err != nil {
			if pathErr, ok := err.(*fs.PathError); ok {
				t.Log("failed at path:", pathErr.Path)
			} else {
				t.Log(err)
			}
		}
	})

	t.Run("error type as check", func(t *testing.T) {
		var pathErr *fs.PathError
		// go 1.26: errors.AsType[*fs.PathError](err)
		if _, err := os.Open("non_existing_file.txt"); err != nil {
			if errors.As(err, &pathErr) {
				t.Log("failed at path:", pathErr.Path)
			} else {
				t.Log(err)
			}
		}
	})
}

// Demo: Json

func TestJsonMarshalTags(t *testing.T) {
	type Person struct {
		ID    int    `json:"id,string"`
		Name  string `json:"name"`
		Level int    `json:"level,omitzero"`
		Desc  string `json:"description,omitempty"`
		// tag:omitzero checks for time.Time IsZero()
		UpdatedBy time.Time `json:"update_by,omitzero"`
	}

	t.Run("json marshal with tags", func(t *testing.T) {
		p := Person{
			ID:        102,
			Name:      "Foo",
			Level:     31,
			Desc:      "A person description",
			UpdatedBy: time.Now(),
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

// Demo: Reg Exp

func TestRegExpMatch(t *testing.T) {
	var emailRegex = regexp.MustCompile(`^[a-z0-9._%+-]+@[a-z0-9.-]+.[a-z]{2,4}$`)

	t.Run("validate email", func(t *testing.T) {
		ok := emailRegex.MatchString("xxxx@google.com")
		t.Log("is matched:", ok)

		ok = emailRegex.MatchString("google.com")
		t.Log("is matched:", ok)
	})
}

func TestRegExpFind(t *testing.T) {
	var idRegex = regexp.MustCompile(`ID:(\d+)`)

	t.Run("find in long content", func(t *testing.T) {
		longContent := "IDs,ID:001,ID:002,ID:003,ID:004,ID:005,ID:006"
		matches := idRegex.FindStringSubmatch(longContent)
		// 这里 id 引用整个 longContent
		// id := matches[1]

		id := strings.Clone(matches[1])
		t.Log("1st matched id:", id)
	})
}
