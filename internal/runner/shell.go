package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/replay/replay/internal/state"
	"github.com/replay/replay/internal/template"
	"github.com/replay/replay/internal/workflow"
)

// Shell executes one or more shell commands.
func Shell(step workflow.Step, store *state.Store) error {
	var commands []string
	var timeoutStr string
	var dirRaw string
	var parallel bool

	if step.Shell != nil {
		if step.Shell.Command != "" {
			commands = append(commands, step.Shell.Command)
		} else if len(step.Shell.Commands) > 0 {
			commands = step.Shell.Commands
		}
		timeoutStr = step.Shell.Timeout
		dirRaw = step.Shell.Dir
		parallel = step.Shell.Parallel
	} else {
		// Shortcuts
		if len(step.Commands) > 0 {
			commands = step.Commands
		} else {
			switch c := step.Command.(type) {
			case string:
				commands = []string{c}
			case []string:
				commands = c
			case []any:
				for _, v := range c {
					commands = append(commands, fmt.Sprintf("%v", v))
				}
			}
		}
		parallel = step.Parallel
		timeoutStr = step.Timeout
		dirRaw = step.Dir
	}

	if len(commands) == 0 {
		return fmt.Errorf("no command provided")
	}

	// Timeout
	var timeout time.Duration
	if timeoutStr != "" {
		d, err := time.ParseDuration(timeoutStr)
		if err != nil {
			return fmt.Errorf("invalid timeout %q: %w", timeoutStr, err)
		}
		timeout = d
	}

	// Working directory
	var dir string
	if dirRaw != "" {
		dir = template.Render(dirRaw, store.All())
	}

	type cmdResult struct {
		stdout string
		stderr string
		err    error
	}

	results := make([]cmdResult, len(commands))

	if parallel && len(commands) > 1 {
		var wg sync.WaitGroup
		for i, cmdRaw := range commands {
			wg.Add(1)
			go func(i int, cmdRaw string) {
				defer wg.Done()
				cmdStr := template.Render(cmdRaw, store.All())

				ctx := context.Background()
				if timeout > 0 {
					var cancel context.CancelFunc
					ctx, cancel = context.WithTimeout(ctx, timeout)
					defer cancel()
				}

				cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
				if dir != "" {
					cmd.Dir = dir
				}

				var stdout, stderr bytes.Buffer
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr

				err := cmd.Run()
				results[i] = cmdResult{
					stdout: stdout.String(),
					stderr: stderr.String(),
					err:    err,
				}

				if err != nil {
					if ctx.Err() == context.DeadlineExceeded {
						results[i].err = fmt.Errorf("command %d failed: timeout after %v", i, timeout)
					} else {
						results[i].err = fmt.Errorf("command %d failed (%v): %s", i, err, strings.TrimSpace(stderr.String()))
					}
				}
			}(i, cmdRaw)
		}
		wg.Wait()
	} else {
		for i, cmdRaw := range commands {
			cmdStr := template.Render(cmdRaw, store.All())

			ctx := context.Background()
			if timeout > 0 {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, timeout)
				defer cancel()
			}

			cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)
			if dir != "" {
				cmd.Dir = dir
			}

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()

			results[i] = cmdResult{
				stdout: stdout.String(),
				stderr: stderr.String(),
				err:    err,
			}

			if err != nil {
				if ctx.Err() == context.DeadlineExceeded {
					results[i].err = fmt.Errorf("command %d failed: timeout after %v", i, timeout)
				} else {
					results[i].err = fmt.Errorf("command %d failed (%v): %s", i, err, strings.TrimSpace(stderr.String()))
				}
				break
			}
		}
	}

	var firstErr error
	for i, r := range results {
		if len(commands) > 1 {
			store.Set(fmt.Sprintf("stdout_%d", i), r.stdout)
			store.Set(fmt.Sprintf("stderr_%d", i), r.stderr)
		}
		if r.err != nil && firstErr == nil {
			firstErr = r.err
		}
	}

	if firstErr != nil {
		return firstErr
	}

	var finalOutput, finalStderr string
	if len(results) > 0 {
		finalOutput = results[len(results)-1].stdout
		finalStderr = results[len(results)-1].stderr
	}

	store.Set("stdout", finalOutput)
	store.Set("stderr", finalStderr)

	if len(step.Extract) > 0 {
		var data interface{}
		errJSON := json.Unmarshal([]byte(finalOutput), &data)

		for varName, path := range step.Extract {
			if errJSON != nil {
				if path == "$" {
					store.Set(varName, finalOutput)
					continue
				}
				return fmt.Errorf("shell stdout is not valid JSON, cannot extract %s: %w", varName, errJSON)
			}

			expr, err := ParseJSONPath(path, step.Name, varName)
			if err != nil {
				return err
			}

			results := expr.Get(data)
			if len(results) == 0 {
				store.Set(varName, nil)
			} else if len(results) == 1 {
				store.Set(varName, results[0])
			} else {
				store.Set(varName, results)
			}
		}
	}

	// Assertions
	ae := NewAssertionEngine(store.All())
	for _, rule := range step.Assert {
		if err := ae.Check(rule, finalOutput); err != nil {
			return fmt.Errorf("assertion failed: %w", err)
		}
	}

	return nil
}
