package structs

import "fmt"

type BitMap struct {
	size int
	data []byte
}

func NewBitMap(size int) BitMap {
	size = (size + 7) / 8 // 向上取整
	if size <= 0 {
		size = 1
	}
	return BitMap{
		data: make([]byte, size),
		size: size,
	}
}

func (b BitMap) Size() int {
	return b.size * 8
}

func (b BitMap) Set(offset int) error {
	index := offset / 8
	if !b.isIndexOutOfRange(index) {
		return fmt.Errorf("offset=%d out of range", offset)
	}

	bitOffset := offset % 8
	b.data[index] |= 1 << bitOffset
	return nil
}

func (b BitMap) Get(offset int) bool {
	index := offset / 8
	if !b.isIndexOutOfRange(index) {
		return false
	}

	bitOffset := offset % 8
	return (b.data[index] & (1 << bitOffset)) != 0
}

func (b BitMap) Reset(offset int) error {
	if ok := b.Get(offset); !ok {
		return nil
	}

	index := offset / 8
	if !b.isIndexOutOfRange(index) {
		return fmt.Errorf("offset=%d out of range", offset)
	}

	bitOffset := offset % 8
	b.data[index] ^= 1 << bitOffset
	return nil
}

func (b BitMap) isIndexOutOfRange(index int) bool {
	return index >= 0 && index < b.size
}
