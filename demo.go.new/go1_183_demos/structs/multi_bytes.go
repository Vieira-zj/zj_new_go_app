package structs

import "io"

type MultiBytes struct {
	data  [][]byte
	index int
	pos   int
}

// 检查 MultiBytes 是否实现了 io.Reader 和 io.Write 接口
var (
	_ io.Reader = (*MultiBytes)(nil)
	_ io.Writer = (*MultiBytes)(nil)
)

func NewMultiBytes(b [][]byte) MultiBytes {
	return MultiBytes{
		data: b,
	}
}

func (mb *MultiBytes) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}

	if mb.index >= len(mb.data) {
		return 0, io.EOF
	}

	n := 0
	for n < len(b) {
		if mb.pos >= len(mb.data[mb.index]) {
			mb.index++
			mb.pos = 0
			if mb.index >= len(mb.data) {
				break
			}
		}

		bytes := mb.data[mb.index]
		cnt := copy(b[n:], bytes[mb.pos:])
		mb.pos += cnt
		n += cnt
	}

	if n == 0 {
		return 0, io.EOF
	}

	return n, nil
}

func (mb *MultiBytes) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	// 创建 clone 副本以避免外部修改影响数据
	clone := make([]byte, len(p))
	copy(clone, p)
	mb.data = append(mb.data, clone)
	return len(clone), nil
}
