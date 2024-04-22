package utils

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"
)

const timeLayout = "2006-01-02 15:04:05"

func FormatDateTime(ti time.Time) string {
	return ti.Format(timeLayout)
}

func IsNil(x any) bool {
	if x == nil {
		return true
	}
	return reflect.ValueOf(x).IsNil()
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

func DeepCopy(src, dest any) error {
	b, err := json.Marshal(src)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(b, &dest); err != nil {
		return err
	}
	return nil
}
