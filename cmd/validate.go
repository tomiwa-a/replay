package cmd

import (
	"fmt"

	"github.com/replay/replay/internal/parser"
	"github.com/replay/replay/internal/validate"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate <workflow.yaml>",
	Short: "Validate a workflow file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wfs, err := parser.LoadFromFile(args[0])
		if err != nil {
			return err
		}

		for _, wf := range wfs {
			if err := validate.Workflow(wf); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "valid workflow: %s (%d steps)\n", wf.Name, len(wf.Steps))
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
}
