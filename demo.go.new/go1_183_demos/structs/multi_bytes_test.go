package structs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// refer: https://github.com/jianghushinian/blog-go-example/blob/main/iox/multi_bytes_test.go

func TestMultiBytesRW(t *testing.T) {
	multiBytes := NewMultiBytes([][]byte{
		[]byte("line one\n"),
	})

	_, err := multiBytes.Write([]byte("line two\n"))
	assert.NoError(t, err)

	b := make([]byte, 12)
	_, err = multiBytes.Read(b)
	assert.NoError(t, err)

	t.Log("bytes:", string(b))
}
