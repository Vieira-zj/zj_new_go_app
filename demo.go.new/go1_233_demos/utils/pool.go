package utils

import (
	"bytes"
	"sync"
)

// Buffer Pool

type BufPool struct {
	internal sync.Pool
}

func NewBufPool() *BufPool {
	return &BufPool{
		internal: sync.Pool{
			New: func() any {
				return new(bytes.Buffer)
			},
		},
	}
}

func (p *BufPool) Put(buf *bytes.Buffer) {
	// 丢弃容量超过 64k 的 buffer
	if buf.Cap() > 64*1024 {
		return
	}
	// Reset 只会把长度变成零, 不会释放 backing array.
	// 如果仍然把这个 buffer 放回 Pool, 进程会继续保留这块大内存, 直到运行时把对象从 Pool 中移除.
	buf.Reset()
	p.internal.Put(buf)
}

func (p *BufPool) Get() *bytes.Buffer {
	buf, ok := p.internal.Get().(*bytes.Buffer)
	if !ok {
		panic("not happen")
	}
	return buf
}

// Request Pool

type Request struct {
	Status  int
	Headers map[string]string
	Body    []byte
}

func (r *Request) Reset() {
	r.Status = 0
	for key := range r.Headers {
		delete(r.Headers, key)
	}
	r.Body = r.Body[:0]
}

// Http Request Pool

type RequestPool struct {
	internal sync.Pool
}

func NewRequestPool() *RequestPool {
	return &RequestPool{
		internal: sync.Pool{
			New: func() any {
				return &Request{
					Headers: make(map[string]string, 8),
					Body:    make([]byte, 0, 1024),
				}
			},
		},
	}
}

func (p *RequestPool) Put(r *Request) {
	r.Reset()
	p.internal.Put(r)
}

func (p *RequestPool) Get() *Request {
	req, ok := p.internal.Get().(*Request)
	if !ok {
		panic("no happen")
	}
	return req
}
