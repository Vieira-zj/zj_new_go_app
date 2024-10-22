package main

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptrace"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Http Client 连接池复用问题

const testUrl = "https://www.baidu.com"

var getConnCount, getDNSCount int32

func TestHttpPing(t *testing.T) {
	err := httpPingByHead()
	assert.NoError(t, err)
}

func TestHttpGet(t *testing.T) {
	t.Run("http get without read body", func(t *testing.T) {
		runHttpGet(t, httpGetWithoutReadBody)
	})

	t.Run("http get with read body", func(t *testing.T) {
		runHttpGet(t, httpGetWithReadBody)
	})
}

func httpPingByHead() error {
	resp, err := http.Head(testUrl)
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusOK {
		fmt.Println("response header:")
		for k, v := range resp.Header {
			fmt.Printf("key=%s, value=%v\n", k, v)
		}
	}

	return nil
}

/*
原理分析

httpclient 每个连接会创建读写协程两个协程, 分别使用 reqch 和 writech 来跟 roundTrip 通信.
上层使用的 response.Body 其实是经过多次封装的, 一次封装的 body 是直接跟 net.conn 进行交互读取, 二次封装的 body 则是加强了 close 和 eof 处理的 bodyEOFSignal.

当未读取 body 就进行 close 时, 会触发 earlyCloseFn() 回调, 看 earlyCloseFn 的函数定义, 在 close 未见 io.EOF 时才调用.
自定义的 earlyCloseFn 方法会给 readLoop 监听的 waitForBodyRead 传入 false, 这样引发 alive 为 false 不能继续循环的接收新请求,
只能是退出调用注册过的 defer 方法, 关闭连接和清理连接池.
*/

func runHttpGet(t *testing.T, httpGetFn func() error) {
	var wg sync.WaitGroup
	limitCh := make(chan struct{}, 5)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		limitCh <- struct{}{}
		go func() {
			defer func() {
				<-limitCh
				wg.Done()
			}()
			if err := httpGetFn(); err != nil {
				t.Log("http get error:", err)
			}
		}()
	}

	wg.Wait()
	close(limitCh)
	t.Logf("getConnCount=%d, getDNSCount=%d", getConnCount, getDNSCount)
}

// httptrace by callbacks.
var testHttpTrace = httptrace.ClientTrace{
	ConnectDone: func(network, addr string, err error) {
		atomic.AddInt32(&getConnCount, 1)
	},
	DNSDone: func(di httptrace.DNSDoneInfo) {
		atomic.AddInt32(&getDNSCount, 1)
	},
}

func httpGetWithoutReadBody() error {
	req, err := http.NewRequest("GET", testUrl, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), &testHttpTrace))
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return err
	}

	// 当未读取 body 就进行 close 时, 会导致请求没有进行连接复用
	defer resp.Body.Close()
	return nil
}

func httpGetWithReadBody() error {
	req, err := http.NewRequest("GET", testUrl, nil)
	if err != nil {
		return err
	}

	req = req.WithContext(httptrace.WithClientTrace(req.Context(), &testHttpTrace))
	resp, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	_, err = io.Copy(io.Discard, resp.Body) // 读取 body
	return err
}
