package utils

import (
	"bufio"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTempFile(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "tmp_test_*")
	assert.NoError(t, err)
	t.Log("create tmp dir:", tmpDir)

	tmpFile, err := os.CreateTemp(tmpDir, "log_*.txt")
	assert.NoError(t, err)
	tmpFile.Close()
	t.Log("create tmp file:", tmpFile.Name())

	err = os.Remove(tmpFile.Name())
	assert.NoError(t, err)
	err = os.Remove(tmpDir)
	assert.NoError(t, err)
}

func TestBufferWriter(t *testing.T) {
	file, err := os.OpenFile("/tmp/test/log.txt", os.O_CREATE|os.O_WRONLY, 0666)
	assert.NoError(t, err)

	bufWriter := bufio.NewWriter(file)
	bufWriter = bufio.NewWriterSize(bufWriter, 5120) // 大于 4096 才会生效

	_, err = bufWriter.Write([]byte{65, 66, 67})
	assert.NoError(t, err)

	_, err = bufWriter.WriteString("Buffered string\n")
	assert.NoError(t, err)

	unflushedBufSize := bufWriter.Buffered()
	t.Logf("Bytes buffered: %d", unflushedBufSize)
	bytesAvailable := bufWriter.Available()
	t.Logf("Available buffer: %d", bytesAvailable)

	err = bufWriter.Flush()
	assert.NoError(t, err)

	bytesAvailable = bufWriter.Available()
	t.Logf("Available buffer: %d", bytesAvailable)
}

func TestReadFileLines(t *testing.T) {
	lines, err := ReadFileLines("/tmp/test/log.txt")
	assert.NoError(t, err)
	t.Log("content:")
	for idx, line := range lines {
		t.Logf("%d: %s", idx, line)
	}
}

func TestReadFileWords(t *testing.T) {
	words, err := ReadFileWords("/tmp/test/log.txt")
	assert.NoError(t, err)
	t.Log("words:")
	for _, word := range words {
		fmt.Print(word + ",")
	}
	fmt.Println()
}

func TestZipArchiveFiles(t *testing.T) {
	err := ZipArchiveFiles("/tmp/test/goc_report", "/tmp/report.zip")
	assert.NoError(t, err)
	t.Log("zip archive done")
}

func TestUnzipArchivedFile(t *testing.T) {
	err := UnzipArchivedFile("/tmp/test/report.zip", "/tmp/test/extract")
	assert.NoError(t, err)
	t.Log("unzip done")
}

func TestGzipCompressFile(t *testing.T) {
	err := GzipCompressFile("/tmp/report.zip", "/tmp/report.zip.gzip")
	assert.NoError(t, err)
	t.Log("gzip compress done")
}

func TestGzipUncompressFile(t *testing.T) {
	err := GzipUncompressFile("/tmp/test/report.zip.gzip", "/tmp/test/report.zip")
	assert.NoError(t, err)
	t.Log("gzip uncompress done")
}
