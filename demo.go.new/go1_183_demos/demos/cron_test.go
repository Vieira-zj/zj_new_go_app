package demos_test

import (
	"testing"
	"time"

	"github.com/robfig/cron/v3"
)

func TestParseStandardCronTimeExp(t *testing.T) {
	// the "standard" cron format for linux
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

func TestParseQuartzCronTimeExp(t *testing.T) {
	// cron format used by the Quartz Scheduler (scheduled jobs in Java)
	parser := cron.NewParser(cron.SecondOptional | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
	for _, exp := range []string{
		"0 0 0/1 * * ?", // every hour
		"10 1 1 * * ?",  // 01:01:10 every day
	} {
		schedule, err := parser.Parse(exp)
		if err != nil {
			t.Fatal(err)
		}

		scheduleNextTime := schedule.Next(time.Now())
		t.Log("schedule next time:", scheduleNextTime.Format(time.DateTime))
	}
}
