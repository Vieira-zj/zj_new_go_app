package unittest

import (
	"context"
	"errors"
	"io/ioutil"
	"runtime"
	"sync"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

//
// cmd: go test -v
//

func TestAdd01(t *testing.T) {
	t.Run("testadd_01", func(t *testing.T) {
		res := Add(0, 1)
		if res != 1 {
			t.Errorf("the result is %d instead of 1", res)
		}
	})

	t.Run("testadd_02", func(t *testing.T) {
		res := Add(1, 0)
		if res != 1 {
			t.Errorf("the result is %d instead of 1", res)
		}
	})
}

func TestAdd02(t *testing.T) {
	data := map[string][]int{
		"case01": {1, 2, 3},
		"case02": {4, 5, 9},
	}

	for name, item := range data {
		t.Run(name, func(t *testing.T) {
			require.Equal(t, item[2], Add(item[0], item[1]))
		})
	}
}

//
// go mock
// mockgen -package unittest -destination mock_test.go io Reader
//

func TestIOReadAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	m := NewMockReader(ctrl)
	m.EXPECT().Read(gomock.Any()).Return(0, errors.New("mock errors"))
	n, err := ioutil.ReadAll(m)
	require.Equal(t, 0, len(n))
	require.Error(t, err)
}

//
// parallel (DATA RACE)
// cmd: go test -v -race
//

func TestIncr01(t *testing.T) {
	var c Counter
	wg := sync.WaitGroup{}
	total := 10
	wg.Add(total)
	for i := 0; i < total; i++ {
		go func() {
			defer wg.Done()
			c.IncrV2()
		}()
	}
	wg.Wait()
	require.Equal(t, total, int(c))
}

func TestIncr02(t *testing.T) {
	var c Counter
	// outer 同步等待并发结果
	t.Run("outer", func(t *testing.T) {
		for i := 0; i < 10; i++ {
			t.Run("inner", func(t *testing.T) {
				t.Parallel()
				c.IncrV2()
			})
		}
		t.Logf("goroutines count: %d\n", runtime.NumGoroutine())
	})
	t.Logf("results: %v\n", c)
}

//
// test returned error value
//

func TestBar(t *testing.T) {
	i1 := impl("i1")
	i2 := impl("i2")
	err := Bar(i1, i2)
	require.Equal(t, "i1", err.Error())
}

//
// test func inputs
// mockgen -package unittest -destination mock_foo_test.go demo.hello/tests/unit.test IFoo
//

func TestFooBar01(t *testing.T) {
	t.Run("testFooBar01", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := NewMockIFoo(ctrl)
		i := 10
		j := 11
		ctx := context.WithValue(context.Background(), impl("k"), impl("v"))
		m.EXPECT().Foo(ctx, i).Return(j, nil)
		b := bar{
			i: m,
		}
		res, err := b.BarV1(ctx, i)
		require.NoError(t, err)
		require.Equal(t, j+1, res)
	})
}

func TestFooBar02(t *testing.T) {
	t.Run("testFooBar02", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		m := NewMockIFoo(ctrl)
		i := 10
		j := 11
		k := impl("k")
		v := impl("v")
		m.EXPECT().Foo(gomock.Any(), i).Do(
			func(ctx context.Context, i int) {
				ret, _ := ctx.Value(k).(string)
				require.Equal(t, v, ret)
			}).Return(j, nil)

		b := bar{
			i: m,
		}
		ctx := context.WithValue(context.Background(), k, v)
		res, err := b.BarV2(ctx, i)
		require.NoError(t, err)
		require.Equal(t, j+1, res)
	})
}
