package cmd

import (
	"fmt"

	"github.com/replay/replay/internal/config"
	"github.com/replay/replay/internal/engine"
	"github.com/replay/replay/internal/parser"
	"github.com/replay/replay/internal/validate"
	"github.com/replay/replay/internal/workflow"
	"github.com/spf13/cobra"
)

var concurrency int
var failFast bool
var profile string
var configFile string
var maxCallDepth int

var runCmd = &cobra.Command{
	Use:           "run <workflow.yaml>...",
	Short:         "Execute one or more workflow files",
	Args:          cobra.MinimumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := loadConfigFile()

		p := &parser.ParserWrapper{}
		e := engine.New(p)
		e.SetDebug(debug)
		e.SetMaxDepth(maxCallDepth)

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
					if cfg != nil {
						if err := config.ApplyConfigToFile(cfg, wf, profile); err != nil {
							return err
						}
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

		pool := engine.NewWorkerPool(concurrency, e)
		pool.Start()

		var allWorkflows []workflow.Workflow
		for _, path := range args {
			wfs, err := parser.LoadFromFile(path)
			if err != nil {
				return err
			}
			allWorkflows = append(allWorkflows, wfs...)
		}

		for i := range allWorkflows {
			wf := &allWorkflows[i]
			if cfg != nil {
				if err := config.ApplyConfigToFile(cfg, wf, profile); err != nil {
					return err
				}
			}
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

		pool.Wait()
		if failFast && pool.HasFailure() {
			return fmt.Errorf("one or more workflows failed")
		}
		return nil
	},
}

func loadConfigFile() *config.ConfigFile {
	path := configFile
	if path == "" {
		path = config.FindConfigFile()
	}
	if path == "" {
		return nil
	}

	cfg, err := config.LoadFromFile(path)
	if err != nil {
		fmt.Printf("Warning: could not load config file: %v\n", err)
		return nil
	}
	return cfg
}

func init() {
	runCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 1, "Number of concurrent workflows")
	runCmd.Flags().BoolVar(&failFast, "fail-fast", false, "Stop execution on first failure")
	runCmd.Flags().StringVar(&profile, "profile", "", "Config profile to use (e.g., dev, staging, prod)")
	runCmd.Flags().StringVar(&configFile, "config", "", "Path to config file (default: replay.yaml)")
	runCmd.Flags().IntVar(&maxCallDepth, "max-call-depth", 100, "Maximum workflow call depth before aborting")
	rootCmd.AddCommand(runCmd)
}
