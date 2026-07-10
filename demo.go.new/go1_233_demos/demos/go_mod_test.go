package demos

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sync/singleflight"
)

func TestRuntimeHooks(t *testing.T) {
	// Runtime Hooks
	// - AddCleanup 负责登记对象不可达后的清理函数
	// - KeepAlive  标出对象必须保持可达的最后位置
	// - SetFinalizer 留给旧代码维护和少数特殊情况

	t.Run("runtime clearup", func(t *testing.T) {
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
	})
}

func TestDecimalCals(t *testing.T) {
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

func TestSingleFlight(t *testing.T) {
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

func TestRunWithSyncTest(t *testing.T) {
	t.Run("run goroutines in synctest", func(t *testing.T) {
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
			const timeout = 5 * time.Second
			ctx, cancel := context.WithTimeout(t.Context(), timeout)
			defer cancel()

			// 刚创建, 超时还没开始, 当然没超时
			time.Sleep(timeout - time.Nanosecond)
			synctest.Wait()
			require.Nil(t, ctx.Err(), "expect nil")

			// 再等 1 纳秒, 触发超时
			time.Sleep(time.Nanosecond)
			synctest.Wait()
			assert.ErrorIs(t, ctx.Err(), context.DeadlineExceeded, "expect DeadlineExceeded")
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

func TestHttpWithSyncTest(t *testing.T) {
	// go test -run ^TestHttpWithSyncTest$ zjin.goapp.demo/demos -v -count=1 -timeout=30s

	synctest.Test(t, func(t *testing.T) {
		// 用 net.Pipe 创建一个 memory mock connection
		srvConn, cliConn := net.Pipe()
		defer cliConn.Close()
		defer srvConn.Close()

		tr := &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return cliConn, nil
			},
			ExpectContinueTimeout: 5 * time.Second,
		}

		// 客户端发起带 "Expect: 100-continue" 的请求
		body := "request body"
		go func() {
			req, _ := http.NewRequest("PUT", "http://test.tld/", strings.NewReader(body))
			req.Header.Set("Expect", "100-continue")
			resp, err := tr.RoundTrip(req)
			assert.Nil(t, err)
			defer resp.Body.Close()
			_, err = io.ReadAll(resp.Body)
			assert.Nil(t, err)
		}()

		// 服务端读取请求头
		req, err := http.ReadRequest(bufio.NewReader(srvConn))
		assert.Nil(t, err)

		// 启动一个 goroutine 读取请求体
		srvGotBody := bytes.Buffer{}
		go func() {
			_, err = io.Copy(&srvGotBody, req.Body)
			assert.Nil(t, err)
		}()

		// 等待所有 goroutine 稳定阻塞
		synctest.Wait()

		// 此时还没发送 100 Continue, 请求体不应该被读取
		assert.Equal(t, srvGotBody.String(), "")

		// 发送 100 Continue
		_, err = srvConn.Write([]byte("HTTP/1.1 100 Continue\r\n\r\n"))
		assert.Nil(t, err)
		synctest.Wait()
		// 现在客户端应该发送了请求体
		assert.Equal(t, srvGotBody.String(), body)

		// 完成请求
		_, err = srvConn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		assert.Nil(t, err)
		// synctest.Test 会自动等待所有 goroutine 退出
	})
}
