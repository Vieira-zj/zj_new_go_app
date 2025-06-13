package utils_test

import (
	"bufio"
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"zjin.goapp.demo/utils"
)

func TestGetBufioInnerReader(t *testing.T) {
	buf := bytes.NewBuffer([]byte("test data"))
	r := bufio.NewReader(buf)
	innerReader := utils.GetBufioInnerReader(r)

	b, err := io.ReadAll(innerReader)
	assert.NoError(t, err)
	t.Log("read data:", string(b))
}
