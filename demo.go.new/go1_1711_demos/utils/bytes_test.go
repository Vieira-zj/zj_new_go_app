package utils_test

import (
	"fmt"
	"go1_1711_demo/utils"
	"math"
	"strings"
	"testing"
	"time"
)

func TestBytesBlob01(t *testing.T) {
	now := uint64(time.Now().Unix())
	hash := uint64(42)
	key := "testkey"
	data := []byte("blob encoding and decoding test")
	blob := utils.WrapBlob(now, hash, key, data)
	t.Log("blob length:", len(blob))

	t.Log("timestamp:", now, utils.ReadTimestampFromBlob(blob))
	t.Log("hash:", hash, utils.ReadHashFromBlob(blob))
	t.Log("key:", key, utils.ReadKeyFromBlob(blob))
	t.Log("data:", string(data), "|", string(utils.ReadDataFromBlob(blob)))

	t.Log("blob length:", len(blob))
}

func TestBytesBlob02(t *testing.T) {
	indexes := []int{0}
	bqueue := make([]byte, 0, 1024)

	for i := 1; i <= 2; i++ {
		now := uint64(time.Now().Unix())
		hash := uint64(42 + int(math.Pow10(i)) + i)
		key := fmt.Sprintf("testkey-%d", i)
		data := []byte(fmt.Sprintf("[%s] blob encoding and decoding test", strings.Repeat("*", i*3)))
		blob := utils.WrapBlob(now, hash, key, data)
		bqueue = append(bqueue, blob...)
		indexes = append(indexes, len(bqueue))
		time.Sleep(time.Second)
	}
	t.Log("bytes length:", len(bqueue))
	fmt.Println()

	for i := 0; i < len(indexes)-1; i++ {
		t.Log("read blob:")
		start := indexes[i]
		end := indexes[i+1]
		blob := bqueue[start:end]
		t.Log("timestamp:", utils.ReadTimestampFromBlob(blob))
		t.Log("hash:", utils.ReadHashFromBlob(blob))
		key := utils.ReadKeyFromBlob(blob)
		t.Log("key:", key)
		data := utils.ReadDataFromBlob(blob)
		t.Log("data:", string(data))
		fmt.Println()
	}
}
