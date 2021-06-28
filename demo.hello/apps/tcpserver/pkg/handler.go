package pkg

import (
	"bufio"
	"context"
	"io"
	"log"
	"net"
	"sync"
	"sync/atomic"
)

// TCPHandler .
type TCPHandler interface {
	Handle(ctx context.Context, conn net.Conn)
	ConnsCount() int
	Close()
}

// EchoHandler .
type EchoHandler struct {
	isClosed   int32
	locker     sync.Mutex
	activeConn map[net.Conn]struct{}
}

// NewEchoHandler .
func NewEchoHandler() *EchoHandler {
	return &EchoHandler{
		isClosed:   1,
		activeConn: make(map[net.Conn]struct{}),
	}
}

// Handle .
func (h *EchoHandler) Handle(ctx context.Context, conn net.Conn) {
	if atomic.LoadInt32(&h.isClosed) == -1 {
		conn.Close()
		return
	}

	h.locker.Lock()
	h.activeConn[conn] = struct{}{}
	h.locker.Unlock()

	reader := bufio.NewReader(conn)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("connection close")
				h.locker.Lock()
				delete(h.activeConn, conn)
				h.locker.Unlock()
			} else {
				log.Println(err)
			}
			return
		}
		conn.Write([]byte(msg))
	}
}

// ConnsCount .
func (h *EchoHandler) ConnsCount() int {
	return len(h.activeConn)
}

// Close .
func (h *EchoHandler) Close() {
	atomic.StoreInt32(&h.isClosed, -1)
	for conn := range h.activeConn {
		conn.Close()
	}
}
