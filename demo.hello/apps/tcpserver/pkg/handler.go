package pkg

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

const (
	ErrReadConnResetByPeer = "server read: connection reset by peer"
	ErrWriteBrokenPipe     = "server write: broken pipe"
	ErrUseOfCLosedConn     = "client: use of closed network connection"
)

// TCPHandle tcp handler interface.
type TCPHandle interface {
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

	h.addActiveConn(conn)
	defer h.delActiveConn(conn)

	if err := writeMessage(conn, "hello, this is from echo"); err != nil {
		log.Println(err)
		return
	}

	connCtx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	// loop: write
	go func() {
		start := time.Now()
		t := time.NewTicker(3 * time.Second)
		defer t.Stop()

		for {
			select {
			case <-ctx.Done():
				log.Println("loop writer exit by server exit")
				return
			case <-connCtx.Done():
				log.Println("loop writer exit by connect close")
				return
			case <-t.C:
				if err := writeMessage(conn, fmt.Sprintf("connect time: %.1fs", time.Since(start).Seconds())); err != nil {
					log.Println(err)
					return
				}
			}
		}
	}()

	// loop: read and write
	reader := bufio.NewReader(conn)
	for {
		// block read until receive data; if eof or connect close from remote, raise error immediately
		msg, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				log.Println("connect close")
			} else {
				log.Println(err)
			}
			return
		}

		if err = writeMessage(conn, "receive: "+msg); err != nil {
			log.Println(err)
			return
		}
	}
}

func (h *EchoHandler) addActiveConn(conn net.Conn) {
	h.locker.Lock()
	h.activeConn[conn] = struct{}{}
	h.locker.Unlock()
}

func (h *EchoHandler) delActiveConn(conn net.Conn) {
	h.locker.Lock()
	if _, ok := h.activeConn[conn]; ok {
		delete(h.activeConn, conn)
	}
	h.locker.Unlock()
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

// Helper

func writeMessage(conn net.Conn, msg string) error {
	if msg[len(msg)-1] != '\n' {
		msg = msg + "\n"
	}
	_, err := conn.Write([]byte(msg))
	return err
}

func isConnInteruptCloseErr(err error) bool {
	return strings.Contains(err.Error(), ErrReadConnResetByPeer) ||
		strings.Contains(err.Error(), ErrWriteBrokenPipe)
}
