package demos

import (
	"bufio"
	"fmt"
	"strings"
	"testing"
)

func TestIOBufScan(t *testing.T) {
	lines := make([]string, 0, 10)
	for i := 0; i < 10; i++ {
		lines = append(lines, fmt.Sprintf("mock line: %d", i))
	}
	content := strings.Join(lines, "\n")

	scanner := bufio.NewScanner(strings.NewReader(content))
	scanner.Split(bufio.ScanLines)

	count := 0
	for scanner.Scan() {
		line := scanner.Text()
		if err := scanner.Err(); err != nil {
			t.Fatal(err)
		}
		count += 1
		t.Log("read line:", line)
	}

	t.Log("lines count:", count)
}
