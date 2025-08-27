package utils_test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/mmap"
)

// read file by mmap

func TestReadByMmap(t *testing.T) {
	// 使用 mmap 的方式, 将磁盘中的文件映射到内存中直接访问, 减少了内核空间与用户空间数据的复制
	const limit = 16

	path := "/tmp/test/output.json"
	mmapf, err := mmap.Open(path) // return ReaderAt, actually bytes in memory
	require.NoError(t, err, "mmap open file failed")
	defer mmapf.Close()

	t.Run("read mmap file once", func(t *testing.T) {
		b := make([]byte, mmapf.Len())
		_, err := mmapf.ReadAt(b, 0)
		assert.NoError(t, err, "mmap read file failed")
		require.True(t, mmapf.Len() > limit, "file content size is incorrect")

		t.Log("total bytes:", mmapf.Len(), len(b))
		t.Log("file content:\n" + string(b[:limit]) + "......" + string(b[len(b)-limit:]))
	})

	t.Run("read mmap file by loop", func(t *testing.T) {
		b, err := ReadAllMmapFile(mmapf)
		assert.NoError(t, err, "mmap read file failed")
		require.True(t, len(b) > limit, "file content size is incorrect")

		t.Log("total bytes:", mmapf.Len(), len(b))
		t.Log("file content:\n" + string(b[:limit]) + "......" + string(b[len(b)-limit:]))
	})
}

func ReadAllMmapFile(f *mmap.ReaderAt) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	b := make([]byte, 4*1024)

	for offset := int64(0); ; {
		n, err := f.ReadAt(b, offset)
		if errors.Is(err, io.EOF) {
			// b := b[:0]
			buf.Write(b[:n])
			break
		}
		if err != nil {
			return nil, fmt.Errorf("mmap read error: %v", err)
		}

		offset += int64(n)
		buf.Write(b[:n])
	}

	return buf.Bytes(), nil
}
