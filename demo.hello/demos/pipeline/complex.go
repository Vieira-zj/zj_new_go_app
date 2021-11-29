package pipeline

import (
	"context"
	"fmt"
	"math/rand"
	"time"
)

// a complex pipeline:
//
//                                  / squarer -> sink
// lines -> lineParser -> splitter -|
//                                   \ sink

// minimalPipelineStage shows the elements that every pipeline stage should have.
// All stages should accept a context for cancellation.
// All stages should return a channel of errors to report any error produced after this function returns.
// All stages should return an error to report any error produced before this function returns.
// Any required input parameters should follow ctx and any required outputs should precede the errors channel.
// Inputs can be ordinary objects (e.g. a list of strings), channels of objects, or gRPC input streams.
// Outputs can be ordinary objects, channels of objects, or gRPC output streams.
func minimalPipelineStage(ctx context.Context) (<-chan error, error) {
	errc := make(chan error, 1)
	go func() {
		defer close(errc)
		// Do something useful here.
	}()
	return errc, nil
}

func splitter(ctx context.Context, in <-chan int64) (<-chan int64, <-chan int64, <-chan error, error) {
	outc1 := make(chan int64)
	outc2 := make(chan int64)
	errc := make(chan error, 1)

	go func() {
		defer close(outc1)
		defer close(outc2)
		defer close(errc)
		for n := range in {
			select {
			case outc1 <- n:
			case <-ctx.Done():
				return
			}
			select {
			case outc2 <- n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return outc1, outc2, errc, nil
}

func squarer(ctx context.Context, in <-chan int64) (<-chan int64, <-chan error, error) {
	outc := make(chan int64)
	errc := make(chan error, 1)

	go func() {
		defer close(outc)
		defer close(errc)
		for n := range in {
			select {
			case outc <- n * n:
			case <-ctx.Done():
				return
			}
		}
	}()
	return outc, errc, nil
}

// RunComplexPipeline runs complex pipeline.
func RunComplexPipeline(base int, lines []string) error {
	fmt.Printf("runComplexPipeline: base=%v, lines=%v\n", base, lines)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var errcList []<-chan error

	// Source pipeline stage.
	linec, errc, err := lineListSource(ctx, lines...)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	// Transformer pipeline stage 1.
	numberc, errc, err := lineParser(ctx, base, linec)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	// Transformer pipeline stage 2.
	numberc1, numberc2, errc, err := splitter(ctx, numberc)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	// Transformer pipeline stage 3.
	numberc3, errc, err := squarer(ctx, numberc1)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	// Sink pipeline stage 1.
	errc, err = sink(ctx, numberc2)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	// Sink pipeline stage 2.
	errc, err = sink(ctx, numberc3)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	fmt.Println("Pipeline started. Waiting for pipeline to complete.")
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
				return
			}
			time.Sleep(time.Second)
		}
	}()

	return outc, errc, nil
}

// RunPipelineWithTimeout runs pipeline with timeout.
func RunPipelineWithTimeout() error {
	fmt.Println("runPipelineWithTimeout")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var errcList []<-chan error

	// Source pipeline stage.
	linec, errc, err := randomNumberSource(ctx, 3)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	// Transformer pipeline stage.
	numberc, errc, err := lineParser(ctx, 10, linec)
	if err != nil {
		return err
	}

	// Sink pipeline stage.
	errc, err = sink(ctx, numberc)
	if err != nil {
		return err
	}
	errcList = append(errcList, errc)

	fmt.Println("Pipeline started. Waiting for pipeline to complete.")

	go func() {
		time.Sleep(time.Duration(10) * time.Second)
		fmt.Println("Cancelling context.")
		cancel()
	}()
	return waitForPipeline(errcList...)
}
