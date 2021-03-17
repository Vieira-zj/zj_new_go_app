package run

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	log "github.com/sirupsen/logrus"
	"golang.org/x/time/rate"

	"demo.grpc/perf/service/utils"
)

/*
Perf Test Worker
*/

type worker struct {
	id     int
	ctx    context.Context
	runner *Runner
	matrix MatrixData
}

func (w *worker) exec(wg *sync.WaitGroup, limiter *rate.Limiter) {
	defer func() {
		w.syncMatrixData()
		wg.Done()
	}()

	log.Debugf("[%d]: Worker start.", w.id)
	w.matrix.Records = []string{}

	go func() {
		tick := time.Tick(time.Duration(w.runner.configs.SyncInterval) * time.Second)
		for {
			select {
			case <-tick:
				w.syncMatrixData()
			case <-w.ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-w.ctx.Done():
			log.Infof("[%d]: Worker cancelled.", w.id)
			return
		default:
		}

		if err := limiter.Wait(w.ctx); err != nil {
			log.Warn(err)
			return
		}
		start := time.Now()
		results, err := w.runner.connect.Get("data") // api test
		w.matrix.Total++
		if err != nil {
			w.matrix.Failed++
			log.Error(err)
		}
		log.Debug("Response result: " + results)

		now := time.Now()
		record := fmt.Sprintf("%s\t%d", utils.TimeFormat(now), (now.Sub(start).Milliseconds()))
		w.matrix.Records = append(w.matrix.Records, record)
	}
}

func (w *worker) syncMatrixData() {
	// log.Debugf("[%d]: Sync worker matrix data: %+v", w.id, w.matrix)
	log.Infof("[%d]: Sync worker matrix data: total=%d, length=%d", w.id, w.matrix.Total, len(w.matrix.Records))

	atomic.AddInt32(&w.runner.Matrix.Total, w.matrix.Total)
	atomic.AddInt32(&w.runner.Matrix.Failed, w.matrix.Failed)
	w.runner.appendTS(w.matrix.Records)

	w.matrix = MatrixData{
		Records: []string{},
	}
}
