package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var tryCmd = &cobra.Command{
	Use:   "try",
	Short: "Try and possibly fail at something",
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("mock error")
	},
}

func init() {
	rootCmd.AddCommand(tryCmd)
}
