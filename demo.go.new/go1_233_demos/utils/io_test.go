package utils_test

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/exp/mmap"

	"zjin.goapp.demo/utils"
)

// case: utils

func TestGetMimeType(t *testing.T) {
	// image mime types:
	// image/png, image/jpeg, image/gif, image/webp and application/octet-stream (unknown)

	t.Run("get image mime type by ext", func(t *testing.T) {
		fpath := "/tmp/test/image.png"
		mimeType := utils.GetMimeTypeByExt(fpath)
		t.Log("mime type from ext:", mimeType)
	})

	t.Run("get image mime type", func(t *testing.T) {
		fpath := "/tmp/test/image"
		mimeType, err := utils.GetMimeType(fpath)
		assert.NoError(t, err)
		t.Log("mime type:", mimeType)
	})
}

// case: buffer read/write

func TestBufferedReadWrite(t *testing.T) {
	const limit = 16

	path := "/tmp/test/output.json"
	f, err := os.Open(path)
	assert.NoError(t, err, "open file failed")
	defer f.Close()

	// buffered read
	r := bufio.NewReaderSize(f, 8*1024) // 8KB buffer
	b, err := io.ReadAll(r)
	assert.NoError(t, err, "file buffered read failed")
	t.Log("file content size:", len(b))

	if len(b) > 100 {
		t.Log("file content:\n" + string(b[:limit]) + "......" + string(b[len(b)-limit:]))
	}

	// buffered write
	outPath := "/tmp/test/out.json"
	outf, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	assert.NoError(t, err, "open output file failed")
	defer outf.Close()

	w := bufio.NewWriterSize(outf, 8*1024)
	n, err := w.Write(b)
	assert.NoError(t, err, "file buffered write failed")
	assert.Equal(t, len(b), n, "write size not match")

	err = w.Flush()
	assert.NoError(t, err, "flush buffered data failed")
	t.Log("write file finished:", outPath)
}

func TestStreamReadByScanner(t *testing.T) {
	path := "/tmp/test/out.json"
	f, err := os.OpenFile(path, os.O_RDONLY, 0644)
	require.NoError(t, err, "open file failed")
	defer f.Close()

	t.Log("scan file:", path)
	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		line := scanner.Text()
		t.Log(line)

		err = scanner.Err()
		require.NoError(t, err, "scan file line failed")
	}
	t.Log("file scan finished")
}

// case: read file by mmap

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
