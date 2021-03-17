package main

import (
	log "github.com/sirupsen/logrus"

	"demo.grpc/perf/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Error(err)
	}
}
