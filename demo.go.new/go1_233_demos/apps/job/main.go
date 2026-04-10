package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

func main() {
	register := registerJobs()

	ctx, cancel := signal.NotifyContext(context.TODO(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	cmd := os.Args[0]
	name := filepath.Base(cmd)
	if job, ok := register[name]; ok {
		job.Run(ctx)
	} else {
		fmt.Println("unknow job name:", name)
	}
}

func registerJobs() map[string]Runer {
	return map[string]Runer{
		"job_foo": NewJobFoo(),
		"job_bar": NewJobBar(),
	}
}
