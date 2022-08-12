package cronjob

import (
	"log"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

func InitCronJob() {
	// add pre-define jobspecs to cronjob.
}

type JobSpec struct {
	Pattern string
	Cmd     func()
}

type CronJob struct {
	client *cron.Cron
	specs  []JobSpec
	lock   sync.Mutex
}

var (
	cronJobInstance *CronJob
	cronJobOnce     sync.Once
)

func NewCronJob() *CronJob {
	cronJobOnce.Do(func() {
		logger := newCronJobLogger()
		c := cron.New(cron.WithSeconds(), cron.WithLogger(logger), cron.WithChain(cron.DelayIfStillRunning(logger), cron.Recover(logger)))
		cronJobInstance = &CronJob{
			client: c,
			specs:  make([]JobSpec, 0, 4),
		}
	})
	return cronJobInstance
}

func (cronjob *CronJob) RegisterCronJobSpec(specs ...JobSpec) {
	cronjob.lock.Lock()
	cronjob.specs = append(cronjob.specs, specs...)
	cronjob.lock.Unlock()
}

func (cronjob *CronJob) Run() {
	for _, spec := range cronjob.specs {
		if len(spec.Pattern) == 0 || spec.Cmd == nil {
			continue
		}
		id, err := cronjob.client.AddFunc(spec.Pattern, spec.Cmd)
		if err != nil {
			log.Fatalln(err)
		}
		log.Printf("Add cron job [%d] success", id)
	}
	cronjob.client.Start()
}

func (cronjob *CronJob) Stop() bool {
	ctx := cronjob.client.Stop()
	select {
	case <-ctx.Done():
		log.Println("cronjob exit")
		return true
	case <-time.After(time.Minute):
		log.Println("cronjob wait, and cancelled")
		return false
	}
}

// Cronjob Logger

type CronJobLogger struct {
}

func newCronJobLogger() cron.Logger {
	return CronJobLogger{}
}

func (clog CronJobLogger) Info(msg string, kvs ...interface{}) {
	log.Printf("msg: %s, values: %v", msg, kvs)
}

func (clog CronJobLogger) Error(err error, msg string, kvs ...interface{}) {
	log.Printf("err: %v, msg: %s, values: %v", err, msg, kvs)
}
