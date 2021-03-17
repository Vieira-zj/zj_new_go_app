package demos

import (
	"context"
	"fmt"
	"strconv"
	"sync"

	"github.com/pkg/errors"
)

// Pipeline Stages
// if upstream failed, current stage (pending for input chan), input chann is closed and go out "for" loop, then return.
// if downstream failed, current stage (pending for output chan), context is cancelled, then return.

func lineListSource(ctx context.Context, lines ...string) (<-chan string, <-chan error, error) {
	// Handle an error that occurs before the goroutine begins.
	if len(lines) == 0 {
		return nil, nil, errors.Errorf("no lines provided")
	}

	out := make(chan string)
	errc := make(chan error, 1) // capacity is 1
	go func() {
		defer close(out)
		defer close(errc)
		for idx, line := range lines {
			if line == "" {
				errc <- errors.Errorf("line %v is empty", idx+1)
				return
			}
			select {
			case out <- line:
				fmt.Println("lineListSource output line:", line)
			case <-ctx.Done():
				fmt.Println("lineListSource canncel")
				return
			}
		}
	}()
	return out, errc, nil
}

func lineParser(ctx context.Context, base int, in <-chan string) (<-chan int64, <-chan error, error) {
	if base < 2 {
		return nil, nil, errors.Errorf("invalid base %v", base)
	}

	out := make(chan int64)
	errc := make(chan error, 1)
	go func() {
		defer close(out)
		defer close(errc)
		for line := range in {
			n, err := strconv.ParseInt(line, base, 64)
			if err != nil {
				errc <- err
				return
			}
			select {
			case out <- n:
				fmt.Println("lineParser output number:", n)
			case <-ctx.Done():
				fmt.Println("lineParser canncel")
				return
			}
		}
	}()
	return out, errc, nil
}

func sink(ctx context.Context, in <-chan int64) (<-chan error, error) {
	errc := make(chan error, 1)
	go func() {
		defer close(errc)
		for n := range in {
			if n >= 100 {
				errc <- errors.Errorf("number %v is too large", n)
				return
			}
			fmt.Printf("sink: %v\n", n)
		}
	}()

	return errc, nil
}

// Helper

// RunSimplePipeline entry for run simple pipeline demo
func RunSimplePipeline(base int, lines []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var errcList []<-chan error
	// Source pipeline stage.
	linec, errc, err := lineListSource(ctx, lines...)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	// Transformer pipeline stage.
	numberc, errc, err := lineParser(ctx, base, linec)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	// Sink pipeline stage.
	errc, err = sink(ctx, numberc)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	return waitForPipeline(errcList...)
}

// waitForPipeline waits for results from all error channels.
func waitForPipeline(errs ...<-chan error) error {
	errc := mergeErrors(errs...)
	for err := range errc {
		if err != nil {
			return err
		}
	}
	return nil
}

// mergeErrors merges multiple channels of errors.
func mergeErrors(cs ...<-chan error) <-chan error {
	var wg sync.WaitGroup
	out := make(chan error, len(cs))

	output := func(c <-chan error) {
		for n := range c {
			out <- n
		}
		wg.Done()
	}

	wg.Add((len(cs)))
	for _, c := range cs {
		go output(c)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}
