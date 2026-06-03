package cmd

import (
	"fmt"

	"github.com/replay/replay/internal/version"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Print())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
