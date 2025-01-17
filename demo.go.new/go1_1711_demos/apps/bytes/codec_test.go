package bytes_test

import (
	"encoding/binary"
	"fmt"
	"go1_1711_demo/apps/bytes"
	"math"
	"strings"
	"testing"
	"time"
)

func TestBinPutUvarint(t *testing.T) {
	bqueue := make([]byte, 0, 16)
	for _, num := range [3]int{100, 1000, 20000} {
		size := bytes.GetIntBytesSize(num)
		t.Logf("number=%d,size=%d", num, size)
		b := make([]byte, size) // 需要提前分配空间
		size = binary.PutUvarint(b, uint64(num))
		bqueue = append(bqueue, b...)
		t.Logf("write: num=%d,size=%d", num, size)
	}

	curIdx := 0
	for i := 0; i < 3; i++ {
		num, size := binary.Uvarint(bqueue[curIdx:])
		t.Logf("read: num=%d,size=%d", num, size)
		curIdx += size
	}
}

func TestWrapAndReadBlob(t *testing.T) {
	now := uint64(time.Now().Unix())
	hash := uint64(42)
	key := "testkey"
	data := []byte("blob encoding and decoding test")
	blob := bytes.WrapBlob(now, hash, key, data)
	t.Log("blob length:", len(blob))

	t.Log("timestamp:", now, bytes.ReadTimestampFromBlob(blob))
	t.Log("hash:", hash, bytes.ReadHashFromBlob(blob))
	t.Log("key:", key, bytes.ReadKeyFromBlob(blob))
	t.Log("data:", string(data), "|", string(bytes.ReadDataFromBlob(blob)))

	t.Log("blob length:", len(blob))
}

func TestReadMultiBlobs(t *testing.T) {
	indexes := []int{0}
	bqueue := make([]byte, 0, 256)

	for i := 1; i <= 2; i++ {
		now := uint64(time.Now().Unix())
		hash := uint64(42 + int(math.Pow10(i)) + i)
		key := fmt.Sprintf("testkey-%d", i)
		data := []byte(fmt.Sprintf("[%s] blob encoding and decoding test", strings.Repeat("*", i*3)))
		blob := bytes.WrapBlob(now, hash, key, data)

		// write blob length
		size := bytes.GetIntBytesSize(len(blob))
		b := make([]byte, size)
		binary.PutUvarint(b, uint64(len(blob)))
		bqueue = append(bqueue, b...)

		// write blob
		bqueue = append(bqueue, blob...)
		indexes = append(indexes, len(bqueue))
		time.Sleep(time.Second)
	}
	t.Log("bytes length:", len(bqueue))
	fmt.Println()

	for i := 0; i < len(indexes)-1; i++ {
		t.Log("read blob:")
		// read blob length
		start := indexes[i]
		len, n := binary.Uvarint(bqueue[start:])

		// read blob
		blob := bqueue[start+n : start+int(len)+1]
		t.Log("timestamp:", bytes.ReadTimestampFromBlob(blob))
		t.Log("hash:", bytes.ReadHashFromBlob(blob))
		key := bytes.ReadKeyFromBlob(blob)
		t.Log("key:", key)
		data := bytes.ReadDataFromBlob(blob)
		t.Log("data:", string(data))
		fmt.Println()
	}
}
