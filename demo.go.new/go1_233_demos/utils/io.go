package utils

import (
	"io"
)

func StreamRead(src io.Reader, dst io.Writer) error {
	buf := make([]byte, 4*1024) // reusable 4k buffer
	_, err := io.CopyBuffer(dst, src, buf)
	return err
}
