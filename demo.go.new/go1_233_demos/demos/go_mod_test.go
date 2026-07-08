package demos

import (
	"fmt"
	"runtime"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"golang.org/x/sync/singleflight"
)

// Runtime Hooks
// - AddCleanup 负责登记对象不可达后的清理函数
// - SetFinalizer 留给旧代码维护和少数特殊情况
// - KeepAlive 标出对象必须保持可达的最后位置

func TestRuntimeClearup(t *testing.T) {
	t.Skip()

	type MockFile struct {
		fd int
	}

	openFile := func(path string) (*MockFile, error) {
		fd, err := syscall.Open(path, syscall.O_RDONLY, 0644)
		if err != nil {
			return nil, err
		}

		f := &MockFile{fd: fd}
		runtime.AddCleanup(f, func(fd int) {
			_ = syscall.Close(fd)
		}, fd)
		return f, nil
	}

	_, err := openFile("/tmp/test/out.json")
	assert.NoError(t, err)
}

func TestDecimalCalculations(t *testing.T) {
	// float64 适合科学计算, decimal/int64 适合财务计算
	t.Run("float calculation", func(t *testing.T) {
		price := 99.995
		taxRate := 0.33
		tax := price * taxRate
		total := price + tax
		t.Logf("float total: %.3f", total)

		t.Log("float equal:", 0.1+0.2 == 0.3)
	})

	t.Run("decimal calculation", func(t *testing.T) {
		price := decimal.NewFromFloat(99.995)
		taxRate := decimal.NewFromFloat(0.13)
		tax := price.Mul(taxRate)
		total := price.Add(tax)
		t.Logf("decimal total: %s", total.StringFixed(3))
	})
}

func TestSingleFlightDemo(t *testing.T) {
	callCount := 0
	mockFetchData := func(key string) (string, error) {
		if len(key) == 0 {
			return "", fmt.Errorf("key is empty")
		}
		callCount++
		fmt.Printf("fetching data for key '%s' from origin (call #%d)...\n", key, callCount)
		time.Sleep(500 * time.Millisecond)
		return "mock_data_for_key|" + key, nil
	}

	const testKey = "singleflight_demo01"
	g := singleflight.Group{}
	wg := sync.WaitGroup{}

	for i := range 5 {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			result, err, shared := g.Do(testKey, func() (any, error) {
				return mockFetchData(testKey)
			})
			if err != nil {
				fmt.Printf("goroutine %d: error fetching data: %v\n", id, err)
				return
			}
			fmt.Printf("goroutine %d: received result: '%v' (shared: %t)\n", id, result, shared)
		}(i)
	}
	wg.Wait()
	t.Logf("total calls to fetch data: %d", callCount)
}
