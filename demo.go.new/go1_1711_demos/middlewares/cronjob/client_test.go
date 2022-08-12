package cronjob

import (
	"fmt"
	"testing"
	"time"
)

// run: go test -timeout 600s -run ^TestCronJob$ go1_1711_demo/middlewares/cronjob -v -count=1
func TestRunCronJob(t *testing.T) {
	jobSpec := JobSpec{
		Pattern: "0 * * * * ?",
		Cmd: func() {
			fmt.Println("Cron job test, run every minutes")
		},
	}

	cronJob := NewCronJob()
	cronJob.RegisterCronJobSpec(jobSpec)
	cronJob.Run()
	defer cronJob.Stop()

	time.Sleep(3 * time.Minute)
	fmt.Println("done")
}
