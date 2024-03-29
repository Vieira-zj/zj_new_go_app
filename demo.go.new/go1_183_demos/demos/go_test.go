package demos_test

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
	"time"
)

// Demo: go test

func TestClearupCase(t *testing.T) {
	t.Cleanup(func() {
		t.Log("case clear")
	})

	if ok := false; !ok {
		t.Fatal("mock fatal")
	}
	t.Log("case run")
}

func MySplit(s, sep string) []string {
	idx := strings.Index(s, sep)
	if idx == -1 {
		return []string{}
	}

	subs := make([]string, 0, 2)
	for idx != -1 {
		subs = append(subs, s[:idx])
		s = s[idx+len(sep):]
		idx = strings.Index(s, sep)
	}
	time.Sleep(time.Second)
	return subs
}

func TestParallelRunCase(t *testing.T) {
	tests := []struct {
		name  string
		input string
		sep   string
		want  []string
	}{
		{"base case", "a:b:c", ":", []string{"a", "b", "c"}},
		{"wrong sep", "a:b:c", ",", []string{"a:b:c"}},
		{"more sep", "abcd", "bc", []string{"a", "d"}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := MySplit(tt.input, tt.sep)
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected:%#v, got:%#v", tt.want, got)
			}
		})
	}
}

// Demo: go stmt

func TestSwitchConds(t *testing.T) {
	getNumberDesc := func(num int) string {
		switch num {
		case 1, 2, 3:
			return "num <= 3"
		case 4, 5:
			return "3 < num <= 5"
		default:
			return "num > 5"
		}
	}

	for i := 2; i < 7; i++ {
		desc := getNumberDesc(i)
		t.Logf("%d: %s", i, desc)
	}
}

func TestTypeAssert(t *testing.T) {
	var s, m any

	t.Run("slice", func(t *testing.T) {
		s = []int{1}
		switch s.(type) {
		case []any:
			t.Log("type: []any")
		case []int:
			t.Log("type: []int")
		default:
			t.Log("unknown type")
		}
	})

	t.Run("map", func(t *testing.T) {
		m = map[string]int{"one": 1}
		switch m.(type) {
		case map[string]any:
			t.Log("type: map[string]any")
		default:
			t.Log("unknown type")
		}
	})
}

// Demo: defer

func TestDeferFn01(t *testing.T) {
	testFn := func() func() {
		t.Log("test fn")
		return func() {
			t.Log("wrapped test fn")
		}
	}

	defer testFn()()
	t.Log("start test defer fn")
	time.Sleep(200 * time.Millisecond)
	t.Log("end test defer fn")
}

type WrappedTest struct {
	t *testing.T
}

func (w *WrappedTest) fn1() *WrappedTest {
	w.t.Log("fn1 invoke")
	return w
}

func (w *WrappedTest) fn2() *WrappedTest {
	w.t.Log("fn2 invoke")
	return w
}

func TestDeferFn02(t *testing.T) {
	s := &WrappedTest{t}
	defer s.fn1().fn2()

	t.Log("start test defer struct fn")
	time.Sleep(200 * time.Millisecond)
	t.Log("end test defer struct fn")
}

// Demo: ref

func TestNilCompare(t *testing.T) {
	var myNil (*byte) = nil

	isNil := true
	if !isNil {
		str := byte(0)
		myNil = &str
	} else {
		t.Log("not init")
	}

	if myNil == nil {
		t.Log("is nil")
	} else {
		t.Log(("is not nil"))
	}
}

func TestRefUpdateForSlice(t *testing.T) {
	update := func(fruits []testFruit) {
		for idx := range fruits {
			fruits[idx].name += "-test"
		}
	}

	fruits := []testFruit{
		{"apple"},
		{"pair"},
	}
	t.Log("fruits:", fruits)
	update(fruits)
	t.Log("new fruits:", fruits)
}

func TestRefUpdateForMap(t *testing.T) {
	update := func(fruits map[string]testFruit) {
		for key, fruit := range fruits {
			fruit.name += "-test"
			fruits[key] = fruit
		}
	}

	fruits := map[string]testFruit{
		"apple": {"apple"},
		"pair":  {"pair"},
	}
	t.Log("fruits:", fruits)
	update(fruits)
	t.Log("new fruits:", fruits)
}

// Demo: goroutine

func TestRecoverFromPanic(t *testing.T) {
	ch := make(chan struct{})

	go func() {
		defer func() {
			if r := recover(); r != nil {
				// print stack from recover
				fmt.Println("recover err:", r)
				fmt.Printf("stack:\n%s", debug.Stack())
			}
			ch <- struct{}{}
		}()
		fmt.Println("goroutine run...")
		time.Sleep(time.Second)
		panic("mock err")
	}()

	fmt.Println("wait...")
	<-ch
	fmt.Println("recover demo done")
}

// Demo: context

func TestContextAfterFunc(t *testing.T) {
	t.Run("run ctx AfterFunc", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		context.AfterFunc(ctx, func() {
			fmt.Println("run ctx clearup")
		})

		t.Log("wait...")
		<-ctx.Done()
		t.Log("cancelled")
	})

	t.Run("stop ctx AfterFunc", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		stop := context.AfterFunc(ctx, func() {
			fmt.Println("run ctx clearup")
		})

		select {
		case <-ctx.Done():
			t.Log("cancelled")
		case <-time.After(200 * time.Millisecond):
			if ok := stop(); ok {
				t.Log("stop AfterFunc")
			}
		}
		t.Log("finish")
	})
}

// Demo: modules

func TestTimeAdd(t *testing.T) {
	duration, err := time.ParseDuration("5m")
	if err != nil {
		t.Fatal(err)
	}
	ti := time.Now().Add(duration)
	t.Log("now after 5m:", ti)

	ti = ti.AddDate(0, 0, 6)
	t.Log("now after 3 days:", ti)
}

func TestGoSlog(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	logger.Debug("text debug level log", "uid", 1002)
	logger.Info("text info level log", "uid", 1002)

	jsonLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	jsonLogger.Debug("json info level log", "uid", 1002)
	jsonLogger.Info("json info level log", "uid", 1002)
}
