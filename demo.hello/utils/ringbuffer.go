package utils

import (
	"errors"
	"fmt"
	"sync"
)

/*
使用场景：异步写日志，允许旧的日志被覆盖。
*/

// RingBuffer .
type RingBuffer struct {
	data        []byte
	size        uint32
	writeCursor uint32
	written     uint32
	locker      *sync.RWMutex
}

// NewRingBuffer .
func NewRingBuffer(size uint32) (*RingBuffer, error) {
	if size <= 0 {
		return nil, fmt.Errorf("Size must be positive")
	}

	return &RingBuffer{
		data:   make([]byte, size),
		size:   size,
		locker: &sync.RWMutex{},
	}, nil
}

// Size returns the size of ring buffer.
func (buf *RingBuffer) Size() uint32 {
	return buf.size
}

// Reset .
func (buf *RingBuffer) Reset() {
	buf.writeCursor = 0
	buf.written = 0
}

// Write writes the bytes with overriding older data if necessary.
func (buf *RingBuffer) Write(inb []byte) (uint32, error) {
	buf.locker.Lock()
	defer buf.locker.Unlock()

	length := uint32(len(inb))
	buf.written += length

	if length > buf.size {
		inb = inb[length-buf.size:]
	}

	// copy in place
	copy(buf.data[buf.writeCursor:], inb)
	// copy from start
	remain := buf.size - buf.writeCursor
	if length > remain {
		copy(buf.data, inb[remain:])
	}

	buf.writeCursor = (buf.writeCursor + uint32(len(inb))) % buf.size
	return length, nil
}

// Read .
func (buf *RingBuffer) Read(seek, length uint32) ([]byte, error) {
	eof := errors.New("eof")
	if seek >= buf.written {
		return nil, eof
	}

	var (
		ret []byte
		err error
	)

	buf.locker.RLock()
	defer buf.locker.RUnlock()

	start := seek % buf.size
	end := start + length
	if start < buf.writeCursor {
		if end >= buf.writeCursor {
			end = buf.writeCursor
			err = eof
		}
		ret = make([]byte, end-start)
		copy(ret, buf.data[start:end])
		return ret, err
	}

	// start >= buf.writeCursor
	if end <= buf.size {
		ret = make([]byte, end-start)
		copy(ret, buf.data[start:end])
		return ret, nil
	}

	// start >= buf.writeCursor && end > buf.size
	end = end - buf.size
	if end >= buf.writeCursor {
		end = buf.writeCursor
		err = eof
	}
	ret = make([]byte, buf.size-start+end)
	copy(ret, buf.data[start:])
	copy(ret[buf.size-start:], buf.data[:end])
	return ret, err
}

// Bytes returns a slice of the bytes written.
func (buf *RingBuffer) Bytes() []byte {
	buf.locker.RLock()
	defer buf.locker.RUnlock()

	switch {
	case buf.written >= buf.size && buf.writeCursor == 0:
		return buf.data
	case buf.written > buf.size:
		outb := make([]byte, buf.size)
		copy(outb, buf.data[buf.writeCursor:])
		copy(outb[buf.size-buf.writeCursor:], buf.data[:buf.writeCursor])
		return outb
	default:
		return buf.data[:buf.writeCursor]
	}
}
