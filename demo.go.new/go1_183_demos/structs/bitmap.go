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
	byteOffset := offset / 8
	if !b.isByteOffsetValid(byteOffset) {
		return fmt.Errorf("invalid offset: %d", byteOffset)
	}

	bitOffset := offset % 8
	b.data[byteOffset] |= 1 << bitOffset
	return nil
}

func (b BitMap) Get(offset int) bool {
	byteOffset := offset / 8
	if !b.isByteOffsetValid(byteOffset) {
		return false
	}

	bitOffset := offset % 8
	return (b.data[byteOffset] & (1 << bitOffset)) != 0
}

func (b BitMap) Reset(offset int) error {
	if ok := b.Get(offset); !ok {
		return nil
	}

	byteOffset := offset / 8
	if !b.isByteOffsetValid(byteOffset) {
		return fmt.Errorf("invalid offset: %d", byteOffset)
	}

	bitOffset := offset % 8
	b.data[byteOffset] ^= 1 << bitOffset
	return nil
}

func (b BitMap) isByteOffsetValid(offset int) bool {
	return offset >= 0 && offset < b.size
}
