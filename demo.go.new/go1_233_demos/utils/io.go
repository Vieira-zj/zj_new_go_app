package utils

import (
	"io"
	// "golang.org/x/exp/mmap"
)

func StreamRead(src io.Reader, dst io.Writer) error {
	buf := make([]byte, 4*1024) // reusable buffer
	_, err := io.CopyBuffer(dst, src, buf)
	return err
}

// 使用 mmap 的方式, 将磁盘中的文件映射到内存中直接访问, 减少了内核空间与用户空间数据的复制
// func ReadByMmap(path string) ([]byte, error) {
// 	f, err := mmap.Open(path)
// 	if err != nil {
// 		return nil, fmt.Errorf("mmap open %s: %w", path, err)
// 	}
// 	defer f.Close()
// 	return io.ReadAll(f)
// }
