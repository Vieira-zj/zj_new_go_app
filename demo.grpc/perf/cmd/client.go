package cmd

import "github.com/spf13/cobra"

var (
	clientCmd = &cobra.Command{
		Use: "client",
	}
)

func init() {
	// TODO:
	rootCmd.AddCommand(clientCmd)
}
