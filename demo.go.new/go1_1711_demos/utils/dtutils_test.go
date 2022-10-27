package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrintFormatTime(t *testing.T) {
	ti := time.UnixMilli(1659541013325)
	t.Log("datetime:", ti)
	t.Log("simple date:", GetSimpleDate(ti))
}

func TestTimeParse(t *testing.T) {
	date := "2022-10-01 00:00:00"
	ti, err := time.Parse(DefaultTimeFormat, date)
	assert.NoError(t, err)
	t.Log("utc:", ti)

	ti, err = time.ParseInLocation(DefaultTimeFormat, date, time.Local)
	assert.NoError(t, err)
	t.Log("cst:", ti)
}

func TestGetStartAndEndMilliOfDate(t *testing.T) {
	now := time.Now()
	for _, dt := range []string{
		"2022-10-01",
		GetSimpleDate(now),
	} {
		start, end, err := GetStartAndEndMilliOfDate(dt)
		assert.NoError(t, err)
		startTime := time.UnixMilli(start)
		endTime := time.UnixMilli(end)
		t.Log("start:", start, startTime)
		t.Log("end", end, endTime)
	}
}

func TestGetDatesInDuration(t *testing.T) {
	res, err := GetDatesInDuration("2022-08-27", "2022-09-03")
	assert.NoError(t, err)
	t.Log(res)
}
