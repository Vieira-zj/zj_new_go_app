package cmd

import (
	"encoding/json"
	"fmt"
	"runtime"

	"demo.grpc/perf/client"
	"demo.grpc/perf/service/run"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile        string
	configurations Configurations

	runCmd = &cobra.Command{
		Use:   "run",
		Short: "run api perf test",
		Long:  "run api perf test with specified paralel number and time seconds",
		Run: func(cmd *cobra.Command, args []string) {
			c, err := json.Marshal(&configurations)
			if err != nil {
				log.Error(err)
				return
			}
			log.Infof("Run perf test configurations: %s", string(c))

			if err := isConfigsValid(); err != nil {
				log.Error(err)
				return
			}
			if err := runPerfTest(); err != nil {
				log.Error(err)
				return
			}
		},
	}
)

func init() {
	runCmd.Flags().StringVarP(&cfgFile, "config", "c", "perf.yml", "config file (default is ./perf.yml)")
	initConfig()
	rootCmd.AddCommand(runCmd)
}

func initConfig() {
	viper.SetConfigType("yml")
	viper.AutomaticEnv()
	viper.SetConfigFile(cfgFile)

	if err := viper.ReadInConfig(); err != nil {
		log.Errorf("Error reading config file, %v", err)
		log.Exit(-1)
	}

	if err := viper.Unmarshal(&configurations); err != nil {
		log.Error(err)
		log.Exit(-1)
	}
}

func isConfigsValid() error {
	if configurations.Runner.RunTime <= 0 {
		return fmt.Errorf("Config runTime [%d] cannot be <= 0", configurations.Runner.RunTime)
	}

	maxParallelNum := 5 * runtime.GOMAXPROCS(-1)
	parallelNum := configurations.Runner.Parallel
	if parallelNum > maxParallelNum {
		return fmt.Errorf("Parallel number [%d] exceed 5*cpu.thread_count [%d]", parallelNum, maxParallelNum)
	}
	return nil
}

func runPerfTest() error {
	conn := client.MockConnect{
		IsRandom:   true,
		Sleep:      100,
		IsError:    false,
		ErrPercent: 1,
	}

	runner := run.NewRunner(&conn, &configurations.Runner)
	if err := runner.Run(); err != nil {
		return err
	}
	log.Infof("Mock api summary: total=%d, failed=%d", conn.Total, conn.Failed)
	return nil
}
