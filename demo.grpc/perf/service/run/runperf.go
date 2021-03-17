package run

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"demo.grpc/perf/client"
	"demo.grpc/perf/service/utils"
)

/*
Perf Test Runner
*/

// Configs perf test runner configs.
type Configs struct {
	Parallel        int
	RunTime         int
	Limit           int
	SyncInterval    int
	OutInterval     int
	FailedThreshold int32
	ReportPath      string `json:"ReportPath,omitempty"`
}

// MatrixData perf test report matrix data.
type MatrixData struct {
	Total   int32
	Failed  int32
	Records []string
}

// Runner runs api perf test cases parallel.
type Runner struct {
	Matrix  MatrixData
	configs *Configs
	connect client.Connection
	locker  *sync.Mutex
}

// NewRunner returns a perf runner instance.
func NewRunner(conn client.Connection, configs *Configs) *Runner {
	return &Runner{
		locker:  &sync.Mutex{},
		connect: conn,
		configs: configs,
	}
}

// Run runs perf test by multiple workers.
func (r *Runner) Run() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(r.configs.RunTime)*time.Second)
	defer cancel()

	go func() {
		tick := time.Tick(time.Duration(r.configs.SyncInterval) * time.Second)
		for {
			select {
			case <-tick:
				if r.Matrix.Failed >= r.configs.FailedThreshold {
					log.Errorf("Failed cases %d, exceed threshold %d.", r.Matrix.Failed, r.configs.FailedThreshold)
					cancel()
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	go func() {
		tick := time.Tick(time.Duration(r.configs.OutInterval) * time.Second)
		for {
			select {
			case <-tick:
				if err := r.reportHandler(); err != nil {
					log.Error(err)
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	const recordTitle = "time\t\t\t\trt(ms)"
	r.Matrix.Records = []string{recordTitle}

	limiter := rate.NewLimiter(rate.Limit(r.configs.Limit), r.configs.Limit)
	wg := sync.WaitGroup{}
	for i := 0; i < r.configs.Parallel; i++ {
		wg.Add(1)
		w := worker{
			id:     i,
			ctx:    ctx,
			runner: r,
		}
		go w.exec(&wg, limiter)
	}
	wg.Wait()

	return r.reportHandler()
}

// appendTS workers sync matrix rt data.
func (r *Runner) appendTS(records []string) {
	r.locker.Lock()
	defer r.locker.Unlock()
	r.Matrix.Records = append(r.Matrix.Records, records...)
}

func (r *Runner) getReportPath() string {
	if len(r.configs.ReportPath) == 0 {
		r.configs.ReportPath = fmt.Sprintf("perf_test_%s.log", utils.TimeFormatWithUnderline(time.Now()))
	}
	return r.configs.ReportPath
}

func (r *Runner) reportHandler() error {
	log.Debug("Output runner matrix data and print summary.")
	if err := r.writeMatrix(); err != nil {
		return err
	}
	return r.printSummary()
}

func (r *Runner) writeMatrix() error {
	r.locker.Lock()
	defer r.locker.Unlock()

	_, err := utils.AppendTextToFile(r.getReportPath(), strings.Join(r.Matrix.Records, "\n"))
	r.Matrix.Records = []string{""}
	return err
}

func (r *Runner) printSummary() error {
	lines, err := utils.ReadFileLines(r.getReportPath())
	if err != nil {
		return err
	}

	sum := 0
	rts := make([]int, 0, len(lines)-1)
	for _, line := range lines[1:] {
		tmp := strings.Split(line, "\t")[1]
		rt, err := strconv.Atoi(strings.TrimRight(tmp, "\n"))
		if err != nil {
			return err
		}
		rts = append(rts, rt)
		sum += rt
	}

	avg := float32(sum) / float32(len(rts))

	sort.Slice(rts, func(i, j int) bool {
		return rts[i] < rts[j]
	})
	log.Debugf("Sorted rts: %v", rts)
	line90 := rts[int(math.Round(float64(len(rts))*0.9)-1)]
	line99 := rts[int(math.Round(float64(len(rts))*0.99)-1)]

	log.Infof("Summary: total=%d, failed=%d, avg=%.2fms, line90=%d, line99=%d", r.Matrix.Total, r.Matrix.Failed, avg, line90, line99)
	return nil
}
