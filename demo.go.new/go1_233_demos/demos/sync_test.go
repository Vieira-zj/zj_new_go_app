package demos

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Context

func TestSubCtxTimeout(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	if ti, ok := ctx.Deadline(); ok {
		t.Log("ctx dead time:", time.Until(ti).Seconds())
	}

	ch := make(chan struct{})
	go func() {
		ctx, cancel = context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if ti, ok := ctx.Deadline(); ok {
			t.Log("sub ctx dead time:", time.Until(ti).Seconds())
		}

		t.Log("goroutine wait")
		<-ctx.Done()

		t.Log("goroutine finish")
		close(ch)
	}()

	<-ch
	t.Log("main finish")
}

// Goroutine

func TestGoroutinesWait(t *testing.T) {
	limit := 3
	wg, lwg := sync.WaitGroup{}, sync.WaitGroup{}
	wg.Add(limit)
	lwg.Add(limit)

	for i := range limit {
		go func(idx int) {
			defer wg.Done()
			log.Printf("goroutine %d start", idx)
			func() {
				defer lwg.Done()
				time.Sleep(time.Duration(i) * time.Second)
			}()

			log.Printf("goroutine %d wait", idx)
			lwg.Wait() // all goroutines run at the same time

			log.Printf("goroutine %d finish", idx)
		}(i)
	}

	wg.Wait()
	t.Log("main finish")
}

// Channel

func TestChannel(t *testing.T) {
	t.Run("close ch from sender", func(t *testing.T) {
		ch := make(chan int, 1)
		go func() {
			for num := range ch {
				fmt.Println("receive:", num)
			}
		}()

		for i := range 10 {
			ch <- i
		}
		close(ch)

		time.Sleep(100 * time.Millisecond) // wait go routine finish
		t.Log("finished")
	})
}

// Sync Once

func TestSyncOnceValue(t *testing.T) {
	calculate := sync.OnceValue(func() int {
		fmt.Println("some complex calculation")
		sum := 0
		for i := range 100_000 {
			sum += i
		}
		return sum
	})

	wg := sync.WaitGroup{}
	for i := range 5 {
		wg.Add(1)
		go func(idx int) {
			t.Logf("run at %d, result: %d", idx, calculate())
			wg.Done()
		}(i)
	}
	wg.Wait()
	t.Log("finished")
}

func TestSyncOnceValues(t *testing.T) {
	readFile := sync.OnceValues(func() ([]byte, error) {
		fmt.Println("read file")
		return os.ReadFile("/tmp/test/output.json")
	})

	wg := sync.WaitGroup{}
	for i := range 3 {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			b, err := readFile()
			if err != nil {
				fmt.Println("read file failed:", err)
				return
			}
			t.Logf("run at %d, read total bytes: %d", idx, len(b))
		}(i)
	}
	wg.Wait()
	t.Log("finished")
}

// Synctest

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
