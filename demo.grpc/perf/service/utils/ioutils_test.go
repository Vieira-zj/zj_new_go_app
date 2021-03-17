package utils

import (
	"testing"
)

func TestReadFileLines(t *testing.T) {
	const path = "/tmp/perf_test.log"
	lines, err := ReadFileLines(path)
	if err != nil {
		t.Fatal(err)
	}

	t.Logf("file lines (%d):\n", len(lines))
	for _, line := range lines {
		t.Logf("%s", line)
	}
}
