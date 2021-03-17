package cmd

import (
	"demo.grpc/perf/service/run"
)

// Configurations perf test run configs.
type Configurations struct {
	Env    string
	Type   string
	Runner run.Configs
}
