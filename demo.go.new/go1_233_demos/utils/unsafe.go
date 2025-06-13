package utils

import (
	"bufio"
	"io"
	"unsafe"
)

type bufioReader struct {
	buf []byte
	rd  io.Reader
}

func GetBufioInnerReader(r *bufio.Reader) io.Reader {
	innerReader := (*bufioReader)(unsafe.Pointer(r))
	return innerReader.rd
}
