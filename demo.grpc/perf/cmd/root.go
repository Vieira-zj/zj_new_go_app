package cmd

import (
	"fmt"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "perf",
	Short: "api perf test",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if configurations.Env != "prod" {
			log.SetLevel(log.InfoLevel)
		} else {
			log.SetLevel(log.WarnLevel)
		}
		log.SetReportCaller(true)
		log.SetFormatter(&log.TextFormatter{
			FullTimestamp: true,
			CallerPrettyfier: func(f *runtime.Frame) (string, string) {
				dirname, filename := filepath.Split(f.File)
				lastelem := filepath.Base(dirname)
				filename = filepath.Join(lastelem, filename)
				return "", fmt.Sprintf("[%s:%d]", filename, f.Line)
			},
		})
	},
}

// Execute runs root command.
func Execute() error {
	return rootCmd.Execute()
}
