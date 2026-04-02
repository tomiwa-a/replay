package cmd

import (
	"github.com/replay/replay/internal/engine"
	"github.com/replay/replay/internal/parser"
	"github.com/replay/replay/internal/validate"
	"github.com/replay/replay/internal/workflow"
	"github.com/spf13/cobra"
)

var concurrency int

var runCmd = &cobra.Command{
	Use:   "run <workflow.yaml>...",
	Short: "Execute one or more workflow files",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		e := engine.New()

		// If concurrency is 1, run sequentially for better output clarity
		if concurrency <= 1 {
			for _, path := range args {
				wfs, err := parser.LoadFromFile(path)
				if err != nil {
					return err
				}

				for i := range wfs {
					wf := &wfs[i]
					if err := validate.Workflow(*wf); err != nil {
						return err
					}

					if err := e.Run(wf); err != nil {
						return err
					}
				}
			}
			return nil
		}

		// Parallel execution
		pool := engine.NewWorkerPool(concurrency, e)
		pool.Start()

		allWorkflows := []workflow.Workflow{}
		for _, path := range args {
			wfs, err := parser.LoadFromFile(path)
			if err != nil {
				return err
			}
			allWorkflows = append(allWorkflows, wfs...)
		}

		for i := range allWorkflows {
			wf := &allWorkflows[i]
			if err := validate.Workflow(*wf); err != nil {
				return err
			}
			pool.Submit(wf)
		}

		pool.Wait()
		return nil
	},
}

func init() {
	runCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 1, "Number of concurrent workflows")
	rootCmd.AddCommand(runCmd)
}
