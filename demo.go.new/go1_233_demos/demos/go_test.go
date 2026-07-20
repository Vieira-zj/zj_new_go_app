package demos

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"

	"testing"
	"testing/quick"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Demo: Struct

type RectanglePb struct {
	Width  *int32
	Height *int32
}

func (r *RectanglePb) GetWidth() int32 {
	if r == nil || r.Width == nil {
		return 0
	}
	return *r.Width
}

func (r *RectanglePb) GetHeight() int32 {
	if r == nil || r.Height == nil {
		return 0
	}
	return *r.Height
}

type RectanglePbWrapped struct {
	*RectanglePb
}

func (w *RectanglePbWrapped) GetArea() int32 {
	if w == nil {
		return 0
	}
	return w.GetWidth() * w.GetHeight()
}

func TestRectanglePbWrapped(t *testing.T) {
	w, h := int32(10), int32(20)
	r := &RectanglePb{
		Width:  &w,
		Height: &h,
	}
	wrapped := &RectanglePbWrapped{
		RectanglePb: r,
	}

	t.Logf("width: %d, height: %d", wrapped.GetWidth(), wrapped.GetHeight())
	t.Log("area:", wrapped.GetArea())
}

// Demo: Common

func TestCalculation(t *testing.T) {
	t.Run("minus uint", func(t *testing.T) {
		a, b := uint32(1), uint32(10)
		t.Log("minus uint32:", int(a-b)) // it will be overflow

		x, y := uint(1), uint(10)
		//nolint:gosec
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
}

func TestVariableOp(t *testing.T) {
	toString := func(s any) string {
		// return s.(string) // panic: interface conversion
		if ret, ok := s.(string); ok {
			return ret
		}
		return "default"
	}

	t.Run("case1: type checking", func(t *testing.T) {
		i := 1
		t.Log("string:", toString(&i))
	})

	t.Run("case2: type checking", func(t *testing.T) {
		type CtxKey string
		ctx := context.TODO()
		val := ctx.Value(CtxKey("id")) // interface{} nil
		t.Log("string:", toString(val))
	})
}

type MyNumbers []string

func (n *MyNumbers) AppendOne(num string) {
	*n = append(*n, num) // here, need to use pointer
}

func TestSliceOp(t *testing.T) {
	t.Run("self type slice append", func(t *testing.T) {
		numbers := MyNumbers([]string{"1", "2"})
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

// Demo: Nil Check

type MyError struct {
	Code    int
	Message string
}

func (e *MyError) Error() string {
	return fmt.Sprintf("code=%d|message=%s", e.Code, e.Message)
}

var _ error = (*MyError)(nil)

func getMyError() error {
	var err *MyError
	return err
}

func TestNilValidate(t *testing.T) {
	t.Run("nil validate", func(t *testing.T) {
		var err1 error
		val := reflect.ValueOf(err1)
		t.Logf("err1 == nil:%v, is_valid:%v", err1 == nil, val.IsValid())

		var err2 *MyError
		val = reflect.ValueOf(err2)
		t.Logf("err1 == nil:%v, is_valid:%v, is_nil:%v, type:%v", err2 == nil, val.IsValid(), val.IsNil(), val.Type())

		// Go 的接口由两部分组成, 类型 (type) + 值 (value)
		err3 := getMyError()
		val = reflect.ValueOf(err3)
		t.Logf("err3 == nil:%v, is_nil:%v, type:%v", err3 == nil, val.IsNil(), val.Type())
	})
}

// Demo: Testing Mods

func TestQuickCheck(t *testing.T) {
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

func TestBuiltInUtils(t *testing.T) {
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

	t.Run("slices grow", func(t *testing.T) {
		s := []int{11, 12, 13}

		limit := 6
		s = slices.Grow(s, limit)

		for i := range limit {
			s = append(s, i+1)
		}
		t.Log("slice after grow:", s)
	})

	t.Run("get array from slices", func(t *testing.T) {
		getArray := func(s []int) ([3]int, error) {
			if len(s) < 3 {
				return [3]int{}, fmt.Errorf("slice too short: %d", len(s))
			}
			// 返回复制 slice 值. 后面修改 s, 不会影响返回的数组
			return [3]int(s[:3]), nil
		}

		s := []int{1, 2, 3, 4, 5}
		result, err := getArray(s)
		assert.NoError(t, err)
		t.Log("array:", result)
	})
}

func TestSlicesSort(t *testing.T) {
	t.Run("sort numbers", func(t *testing.T) {
		s := []uint32{21, 22, 1, 2, 3, 4}
		slices.SortFunc(s, func(a, b uint32) int {
			// return int(a - b) // it will be overflow
			return cmp.Compare(a, b)
		})
		t.Log("sorted uint32 slice:", s)
	})

	t.Run("multiple fields sort", func(t *testing.T) {
		type User struct {
			Name  string
			Score int
		}
		users := []User{
			{"Carol", 90},
			{"Alice", 95},
			{"Bob", 80},
		}
		slices.SortFunc(users, func(a, b User) int {
			return cmp.Or(
				cmp.Compare(a.Score, b.Score),
				strings.Compare(a.Name, b.Name),
			)
		})
		t.Log("sort users:", users)
	})

	t.Run("sorted iter.Seq", func(t *testing.T) {
		scores := map[string]int{
			"Carol": 90,
			"Alice": 95,
			"Bob":   80,
		}
		names := slices.Sorted(maps.Keys(scores))
		t.Log("names:", names)
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
	t.Run("os tmp dir", func(t *testing.T) {
		t.Log("tmp dir:", os.TempDir())
	})

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

func TestFilePathUtil(t *testing.T) {
	t.Run("search files in dir", func(t *testing.T) {
		// glob 不支持多级目录
		dir := "/Users/jinzheng/Downloads/tmps/"
		names, err := filepath.Glob(dir + "*.yml")
		assert.NoError(t, err)

		t.Log("yml files:")
		for _, name := range names {
			t.Log(name)
		}
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

	t.Run("unwrap error", func(t *testing.T) {
		if _, err := os.Open("non_existing_file.txt"); err != nil {
			if pathErr, ok := err.(*fs.PathError); ok {
				t.Log("failed at path:", pathErr.Path)
			} else {
				t.Log(err)
			}
		}
	})

	t.Run("unwrap error by as", func(t *testing.T) {
		// check error type, if matched, then unwrap error and save in 'pathErr' (impl by Unwrap())
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

	t.Run("errors join", func(t *testing.T) {
		err := errors.Join(io.ErrClosedPipe, io.ErrUnexpectedEOF, context.Canceled)
		t.Log("joined error:", err)
		t.Log("is cannceled error:", errors.Is(err, context.Canceled))
	})
}
