package utils_test

import (
	"testing"
	"time"

	"demo.apps/utils"
	"github.com/stretchr/testify/assert"
)

func TestGetNextWorkDayAfterDays(t *testing.T) {
	for _, tc := range []struct {
		desc   string
		input  string
		expect string
	}{
		{"test not weekend", "2024-09-03", "2024-09-06"},
		{"test weekend 1", "2024-09-04", "2024-09-09"},
		{"test weekend 2", "2024-09-05", "2024-09-10"},
		{"test weekend 3", "2024-09-06", "2024-09-11"},
		{"test weekend 4", "2024-11-09", "2024-11-14"},
	} {
		t.Run(tc.desc, func(t *testing.T) {
			ti, err := time.Parse(time.DateOnly, tc.input)
			assert.NoError(t, err)
			actual := utils.GetNextWorkDateAfterDays(ti, 3)
			assert.Equal(t, tc.expect, actual.Format(time.DateOnly))
		})
	}
}
