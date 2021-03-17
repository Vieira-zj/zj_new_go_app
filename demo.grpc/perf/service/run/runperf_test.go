package run

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"demo.grpc/perf/client"
)

func TestTimer(t *testing.T) {
	now := time.Now()
	year, month, day := now.Date()
	hour, min, sec := now.Clock()
	t.Log(fmt.Sprintf("%d-%d-%d %d:%d:%d", year, month, day, hour, min, sec))
	t.Log(now.Format("2006-01-02 15:04:05"))

	time.Sleep(time.Second)
	t.Log(time.Now().Sub(now).Milliseconds())
}

func TestTicker(t *testing.T) {
	done := make(chan struct{})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	go func() {
		tick := time.Tick(2 * time.Second)
		for {
			select {
			case <-tick:
				t.Log("work...")
			case <-ctx.Done():
				t.Log("finished")
				close(done)
				return
			}
		}
	}()

	<-done
	log.Println("ticker test done")
}

type myInt int

var myValue myInt

func TestContext(t *testing.T) {
	runTime := 2
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(runTime)*time.Second)
	defer cancel()

	ctx = context.WithValue(ctx, myValue, 9)

	wg := sync.WaitGroup{}
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, id int) {
			defer wg.Done()

			t.Logf("worker [%d] start\n", id)
			for {
				select {
				case <-ctx.Done():
					t.Log("finished")
					return
				default:
				}
				t.Log("worker is run ...")
				v := ctx.Value(myValue).(int)
				t.Log("context value:", v)
				time.Sleep(time.Second)
			}
		}(ctx, &wg, i)
	}
	wg.Wait()
	t.Log("context test done")
}

func TestPerfRun(t *testing.T) {
	conn := client.MockConnect{
		IsRandom:   true,
		Sleep:      100,
		IsError:    false,
		ErrPercent: 5,
	}

	configs := Configs{
		Parallel:        2,
		RunTime:         25,
		Limit:           50,
		SyncInterval:    5,
		OutInterval:     10,
		FailedThreshold: 10,
	}

	runner := Runner{
		locker:  &sync.Mutex{},
		connect: &conn,
		configs: &configs,
	}

	if err := runner.Run(); err != nil {
		t.Fatal(err)
	}
	t.Logf("mock api invoked: total=%d, failed=%d\n", conn.Total, conn.Failed)
}
