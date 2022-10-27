package utils

import (
	"fmt"
	"time"
)

const DefaultTimeFormat = "2006-01-02 15:04:05"

func GetPreTimeRange(preTime time.Duration) (int64, int64) {
	now := time.Now()
	end := now.UnixMilli()
	start := now.Add(-preTime).UnixMilli()
	return start, end
}

// GetStartAndEndMilliOfDate dt must be "2006-01-02" but not "2006-1-2".
func GetStartAndEndMilliOfDate(dt string) (int64, int64, error) {
	date := fmt.Sprintf("%s 00:00:00", dt)
	ti, err := time.ParseInLocation(DefaultTimeFormat, date, time.Local)
	if err != nil {
		return -1, -1, err
	}
	fmt.Println(ti)
	return ti.UnixMilli(), ti.Add(24*time.Hour - time.Second).UnixMilli(), nil
}

func GetDatesInDuration(start, end string) ([]string, error) {
	startTime, err := time.ParseInLocation(DefaultTimeFormat, fmt.Sprintf("%s 00:00:00", start), time.Local)
	if err != nil {
		return nil, err
	}
	endTime, err := time.ParseInLocation(DefaultTimeFormat, fmt.Sprintf("%s 00:00:00", end), time.Local)
	if err != nil {
		return nil, err
	}

	subDays := int(endTime.Sub(startTime).Hours() / 24)
	retDates := make([]string, 0, subDays)
	for i := 0; i <= subDays; i++ {
		tmpTime := startTime.Add(time.Duration(i) * 24 * time.Hour)
		retDates = append(retDates, GetSimpleDate(tmpTime))
	}
	return retDates, nil
}

func GetSimpleDate(t time.Time) string {
	return fmt.Sprintf("%d-%02d-%02d", t.Year(), t.Month(), t.Day())
}
