package pipeline

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/pkg/errors"
)

//
// Pipeline Stages:
// lineListSource -> lineParser -> sink
//
// if upstream failed, current stage (pending for input chan), input chan is closed and go out "for" loop, then return.
// if downstream failed, current stage (pending for output chan), context is cancelled, then return.
//

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

//
// Run simple pipeline
//

// RunSimplePipeline entry for run simple pipeline demo.
func RunSimplePipeline(base int, lines []string) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var errcList []<-chan error
	// stage1: Source pipeline
	linec, errc, err := lineListSource(ctx, lines...)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	// stage2: Transformer pipeline
	numberc, errc, err := lineParser(ctx, base, linec)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	// stage3: Sink pipeline
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

// mergeErrors merges multiple channels of errors into one output channel.
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

//
// Run simple pipeline with timeout
//

// RunPipelineWithTimeout runs pipeline with timeout.
func RunPipelineWithTimeout(timeout int) error {
	fmt.Println("runPipelineWithTimeout")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var errcList []<-chan error

	// stage1: Source pipeline
	linec, errc, err := randomNumberSource(ctx, time.Now().Unix())
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	// stage2: Transformer pipeline
	numberc, errc, err := lineParser(ctx, 10, linec)
	if err != nil {
		return err
	}

	// stage3: Sink pipeline
	errc, err = sink(ctx, numberc)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)
	fmt.Println("Pipeline started. Waiting for pipeline to complete.")

	go func() {
		time.Sleep(time.Duration(timeout) * time.Second)
		fmt.Println("Cancelling context")
		cancel()
	}()
	return waitForPipeline(errcList...)
}

func randomNumberSource(ctx context.Context, seed int64) (<-chan string, <-chan error, error) {
	outc := make(chan string)
	errc := make(chan error, 1)
	random := rand.New(rand.NewSource(seed))

	go func() {
		defer close(outc)
		defer close(errc)
		for {
			n := random.Intn(100)
			line := fmt.Sprintf("%v", n)
			select {
			case outc <- line:
			case <-ctx.Done():
				fmt.Println("Source exit:", ctx.Err())
				return
			}
			time.Sleep(time.Second)
		}
	}()
	return outc, errc, nil
}
