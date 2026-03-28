package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replay executes declarative workflow files",
}

func Execute() error {
	return rootCmd.Execute()
}
