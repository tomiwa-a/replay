package cmd

import (
	"fmt"

	"github.com/replay/replay/internal/engine"
	"github.com/replay/replay/internal/parser"
	"github.com/replay/replay/internal/validate"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run <workflow.yaml>",
	Short: "Execute a workflow file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		wf, err := parser.LoadFromFile(args[0])
		if err != nil {
			return err
		}

		if err := validate.Workflow(wf); err != nil {
			return err
		}

		fmt.Printf("Validating workflow %q... OK\n", wf.Name)

		e := engine.New()
		if err := e.Run(wf); err != nil {
			return err
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(runCmd)
}
