package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTimeCommon(t *testing.T) {
	d := time.Duration(time.Second)
	t.Log("milli sec:", d.Milliseconds())

	// since and until
	start := time.Now()
	time.Sleep(3 * time.Second)
	t.Logf("time since: %.2fs\n", time.Since(start).Seconds())

	end := start.AddDate(0, 0, 1)
	t.Logf("time until: %.2fh\n", time.Until(end).Hours())

	// parse and format
	const (
		layoutISO = "2006-01-02"
		layoutUS  = "January 2, 2006"
	)
	date := "2012-08-09"
	ti, err := time.ParseInLocation(layoutISO, date, time.Local)
	assert.NoError(t, err)
	t.Log(ti) // 时区 CST, China Standard Time
	t.Log("us time:", ti.Format(layoutUS))

	// location
	now := time.Now()
	loc, err := time.LoadLocation("UTC")
	assert.NoError(t, err)
	t.Log(now.In(loc))

	loc, err = time.LoadLocation("Asia/ShangHai")
	assert.NoError(t, err)
	t.Log(now.In(loc))
}

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
