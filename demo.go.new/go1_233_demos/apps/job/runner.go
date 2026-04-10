package main

import (
	"context"
	"fmt"
	"time"
)

type Runer interface {
	Run(context.Context)
}

// Job Foo

var _ Runer = (*JobFoo)(nil)

type JobFoo struct{}

func NewJobFoo() *JobFoo {
	return &JobFoo{}
}

func (j *JobFoo) Run(ctx context.Context) {
	fmt.Println("job foo start")

	t := time.NewTicker(time.Second)
	defer t.Stop()

	for range 3 {
		select {
		case <-t.C:
			fmt.Println("job foo is running")
		case <-ctx.Done():
			fmt.Println("job foo is canceled")
		}
	}
	fmt.Println("job foo finish")
}

// Job Bar

var _ Runer = (*JobBar)(nil)

type JobBar struct{}

func NewJobBar() *JobBar {
	return &JobBar{}
}

func (j *JobBar) Run(ctx context.Context) {
	fmt.Println("job bar start")

	t := time.NewTicker(time.Second)
	defer t.Stop()

	for range 2 {
		select {
		case <-t.C:
			fmt.Println("job bar is running")
		case <-ctx.Done():
			fmt.Println("job bar is canceled")
		}
	}
	fmt.Println("job bar finish")
}
