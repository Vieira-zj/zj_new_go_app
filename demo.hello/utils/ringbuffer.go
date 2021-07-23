package utils

import "fmt"

// RingBuffer .
type RingBuffer struct {
	data        []byte
	size        uint32
	writeCursor uint32
	written     uint32
}

// NewRingBuffer .
func NewRingBuffer(size uint32) (*RingBuffer, error) {
	if size <= 0 {
		return nil, fmt.Errorf("Size must be positive")
	}

	return &RingBuffer{
		data: make([]byte, size),
		size: size,
	}, nil
}

// Size returns the size of ring buffer.
func (buf *RingBuffer) Size() int {
	return int(buf.size)
}

// Reset .
func (buf *RingBuffer) Reset() {
	buf.writeCursor = 0
	buf.written = 0
}

// Write writes the bytes with overriding older data if necessary.
func (buf *RingBuffer) Write(inb []byte) (uint32, error) {
	n := uint32(len(inb))
	buf.written += n

	if n > buf.size {
		inb = inb[n-buf.size:]
	}

	// copy in place
	remain := buf.size - buf.writeCursor
	copy(buf.data[buf.writeCursor:], inb)
	// copy from start
	if n > remain {
		copy(buf.data, inb[remain:])
	}

	buf.writeCursor = (buf.writeCursor + uint32(len(inb))) % buf.size
	return n, nil
}

// Bytes .
func (buf *RingBuffer) Bytes() []byte {
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
