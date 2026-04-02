package cmd

import "github.com/spf13/cobra"

var rootCmd = &cobra.Command{
	Use:   "replay",
	Short: "Replay executes declarative E2E testing workflows",
	Long: `Replay is a standalone, CLI-based execution engine for declarative E2E testing workflows.
It combines HTTP calls, native Database operations (PostgreSQL/Redis), and Shell commands
into unified, stateful test suites.

To enable IDE autocomplete and AI awareness, add this to the top of your YAML files:
# yaml-language-server: $schema=https://raw.githubusercontent.com/tomiwa-a/replay/main/schema.json`,
}

var debug bool

func init() {
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug logging for HTTP and other runners")
}

func Execute() error {
	return rootCmd.Execute()
}
