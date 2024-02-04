package demos

import (
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

func TestParseCronTimeExp(t *testing.T) {
	for _, exp := range []string{
		"0 * * * *",   // every hour
		"45 23 * * 1", // 23:45 every monday
	} {
		schedule, err := cron.ParseStandard(exp)
		if err != nil {
			t.Fatal(err)
		}

		scheduleNextTime := schedule.Next(time.Now())
		t.Log("schedule next time:", scheduleNextTime.Format(time.DateTime))
	}
}
