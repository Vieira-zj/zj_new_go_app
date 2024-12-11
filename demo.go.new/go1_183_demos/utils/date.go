package utils

import "time"

func IsWeekend(ti time.Time) bool {
	day := ti.Weekday()
	return day == time.Saturday || day == time.Sunday
}

func IsDateEqual(src, dst time.Time) bool {
	src = src.Truncate(24 * time.Hour)
	dst = dst.Truncate(24 * time.Hour)
	return src.Equal(dst)
}

// GetNextWorkDateAfterDays get next working date after days (include current day).
func GetNextWorkDateAfterDays(date time.Time, days uint32) time.Time {
	for IsWeekend(date) {
		date = date.AddDate(0, 0, 1)
	}

	for i := 0; i < int(days); {
		date = date.AddDate(0, 0, 1)
		if !IsWeekend(date) {
			i++
		}
	}
	return date
}
