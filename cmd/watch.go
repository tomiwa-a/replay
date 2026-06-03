package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/replay/replay/internal/config"
	"github.com/replay/replay/internal/engine"
	"github.com/replay/replay/internal/parser"
	"github.com/replay/replay/internal/validate"
	"github.com/spf13/cobra"
)

var watchDebounce time.Duration

var watchCmd = &cobra.Command{
	Use:           "watch <workflow.yaml>...",
	Short:         "Watch workflow files and re-run on changes",
	Args:          cobra.MinimumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("failed to create watcher: %w", err)
		}
		defer watcher.Close()

		cfg := loadConfigFile()

		absPaths := make([]string, len(args))
		for i, arg := range args {
			abs, err := filepath.Abs(arg)
			if err != nil {
				return fmt.Errorf("failed to resolve path %q: %w", arg, err)
			}
			absPaths[i] = abs
		}

		watchDirs := make(map[string]bool)
		for _, p := range absPaths {
			dir := filepath.Dir(p)
			if !watchDirs[dir] {
				watchDirs[dir] = true
				if err := watcher.Add(dir); err != nil {
					return fmt.Errorf("failed to watch directory %q: %w", dir, err)
				}
			}
		}

		fmt.Printf("Watching %d file(s) for changes (debounce: %v)...\n", len(absPaths), watchDebounce)
		for _, p := range absPaths {
			fmt.Printf("  %s\n", filepath.Base(p))
		}
		fmt.Println("Press Ctrl+C to stop.")

		if err := runWorkflows(cfg, absPaths); err != nil {
			fmt.Printf("Error: %v\n", err)
		}

		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

		var debounceTimer *time.Timer

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				if event.Op&(fsnotify.Write|fsnotify.Create) == 0 {
					continue
				}
				if filepath.Ext(event.Name) != ".yaml" && filepath.Ext(event.Name) != ".yml" {
					continue
				}
				if !isWatched(event.Name, absPaths) {
					continue
				}

				if debounceTimer != nil {
					debounceTimer.Stop()
				}
				debounceTimer = time.AfterFunc(watchDebounce, func() {
					fmt.Printf("\n--- File changed: %s ---\n", filepath.Base(event.Name))
					if err := runWorkflows(cfg, absPaths); err != nil {
						fmt.Printf("Error: %v\n", err)
					}
					fmt.Println("Watching for changes...")
				})

			case err, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				fmt.Printf("Watcher error: %v\n", err)

			case <-sigChan:
				fmt.Println("\nStopping watcher.")
				return nil
			}
		}
	},
}

func runWorkflows(cfg *config.ConfigFile, paths []string) error {
	p := &parser.ParserWrapper{}
	e := engine.New(p)
	e.SetDebug(debug)

	for _, path := range paths {
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
				return err
			}
		}
	}
	return nil
}

func isWatched(changedPath string, watchedPaths []string) bool {
	changedAbs, err := filepath.Abs(changedPath)
	if err != nil {
		return false
	}
	for _, wp := range watchedPaths {
		if wp == changedAbs {
			return true
		}
	}
	return false
}

func init() {
	watchCmd.Flags().DurationVar(&watchDebounce, "debounce", 500*time.Millisecond, "Debounce duration for file changes")
	watchCmd.Flags().IntVarP(&concurrency, "concurrency", "c", 1, "Number of concurrent workflows")
	watchCmd.Flags().BoolVar(&failFast, "fail-fast", false, "Stop execution on first failure")
	watchCmd.Flags().StringVar(&profile, "profile", "", "Config profile to use")
	watchCmd.Flags().StringVar(&configFile, "config", "", "Path to config file")
	rootCmd.AddCommand(watchCmd)
}
