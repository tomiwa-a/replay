package cmd

import (
	"fmt"

	"github.com/replay/replay/internal/engine"
	"github.com/replay/replay/internal/parser"
	"github.com/replay/replay/internal/validate"
	"github.com/replay/replay/internal/workflow"
	"github.com/spf13/cobra"
)

var concurrency int
var failFast bool

var runCmd = &cobra.Command{
	Use:           "run <workflow.yaml>...",
	Short:         "Execute one or more workflow files",
	Args:          cobra.MinimumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		p := &parser.ParserWrapper{}
		e := engine.New(p)
		e.SetDebug(debug)

		// If concurrency is 1, run sequentially for better output clarity
		if concurrency <= 1 {
			for _, path := range args {
				wfs, err := parser.LoadFromFile(path)
				if err != nil {
					return err
				}

				for i := range wfs {
					wf := &wfs[i]
					if debug {
						wf.Config.HTTP.Debug = true
					}
					if err := parser.ResolveIncludes(wf); err != nil {
						return err
					}
					if err := validate.Workflow(*wf); err != nil {
						return err
					}

					if err := e.Run(wf); err != nil {
						if failFast {
							return err
						}
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
			if err := parser.ResolveIncludes(wf); err != nil {
				return err
			}
			if err := validate.Workflow(*wf); err != nil {
				return err
			}
			if failFast && pool.HasFailure() {
				continue
			}
			pool.Submit(wf)
		}

		for i := range allWorkflows {
			wf := &allWorkflows[i]
			if err := validate.Workflow(*wf); err != nil {
				return err
			}
			if failFast && pool.HasFailure() {
				continue
			}
			pool.Submit(wf)
		}

		pool.Wait()
		if failFast && pool.HasFailure() {
			return fmt.Errorf("one or more workflows failed")
		}
		return nil
	},
}

func init() {
	runCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 1, "Number of concurrent workflows")
	runCmd.Flags().BoolVar(&failFast, "fail-fast", false, "Stop execution on first failure")
	rootCmd.AddCommand(runCmd)
}
