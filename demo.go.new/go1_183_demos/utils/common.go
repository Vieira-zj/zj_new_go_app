package utils

import (
	"fmt"
	"time"
)

const timeLayout = "2006-01-02 15:04:05"

func FormatDateTime(ti time.Time) string {
	return ti.Format(timeLayout)
}

func TrackTime() func() {
	start := time.Now()
	return func() {
		elapsed := time.Since(start)
		fmt.Printf("elapsed: %.2fs\n", elapsed.Seconds())
	}
}

func DelFirstNItemsOfSlice(s []any /* will change input slice */, n int) ([]any, error) {
	if n >= len(s) {
		return nil, fmt.Errorf("n must be less than length of input slice")
	}

	m := copy(s, s[n:])
	for i := m; i < len(s); i++ {
		s[i] = nil // avoid memory leaks
	}

	s = s[:m] // reset length
	return s, nil
}
