package cmd

import "github.com/spf13/cobra"

var (
	serverCmd = &cobra.Command{
		Use: "server",
	}
)

func init() {
	// TODO:
	rootCmd.AddCommand(serverCmd)
}
