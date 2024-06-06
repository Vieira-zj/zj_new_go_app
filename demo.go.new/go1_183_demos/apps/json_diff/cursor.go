package jsondiff

import (
	"strconv"
)

const sep byte = '/'

type cursor struct {
	buf        []byte
	sepIndexes []int
}

func (c cursor) string() string {
	return string(c.buf)
}

func (c cursor) isRoot() bool {
	return len(c.buf) == 0
}

func (c *cursor) appendKey(key string) {
	c.buf = append(c.buf, sep)
	c.sepIndexes = append(c.sepIndexes, len(c.buf)-1)
	c.buf = append(c.buf, key...)
}

func (c *cursor) appendIndex(idx int) {
	c.buf = append(c.buf, sep)
	c.sepIndexes = append(c.sepIndexes, len(c.buf)-1)
	c.buf = strconv.AppendInt(c.buf, int64(idx), 10)
}

// rollback resets cursor to last sep index.
func (c *cursor) rollback() {
	lastIdx := len(c.sepIndexes) - 1
	idx := c.sepIndexes[lastIdx]
	c.buf = c.buf[:idx]
	c.sepIndexes = c.sepIndexes[:lastIdx]
}

func (c *cursor) reset() {
	c.buf = c.buf[:0]
	c.sepIndexes = c.sepIndexes[:0]
}
