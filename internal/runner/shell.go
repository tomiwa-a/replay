package runner

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/replay/replay/internal/state"
	"github.com/replay/replay/internal/template"
	"github.com/replay/replay/internal/workflow"

	"github.com/ohler55/ojg/jp"
)

// Shell executes one or more shell commands.
func Shell(step workflow.Step, store *state.Store) error {
	if step.Shell == nil {
		return fmt.Errorf("shell configuration is missing")
	}

	var commands []string
	if step.Shell.Command != "" {
		commands = append(commands, step.Shell.Command)
	} else if len(step.Shell.Commands) > 0 {
		commands = step.Shell.Commands
	}

	if len(commands) == 0 {
		return fmt.Errorf("no command provided")
	}

	// Timeout
	var timeout time.Duration
	if step.Shell.Timeout != "" {
		d, err := time.ParseDuration(step.Shell.Timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout %q: %w", step.Shell.Timeout, err)
		}
		timeout = d
	}

	// Working directory
	var dir string
	if step.Shell.Dir != "" {
		dir = template.Render(step.Shell.Dir, store.All())
	}

	var finalOutput string
	var finalStderr string

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

		output := stdout.String()
		errStr := stderr.String()

		finalOutput = output
		finalStderr = errStr

		if len(commands) > 1 {
			store.Set(fmt.Sprintf("stdout_%d", i), output)
			store.Set(fmt.Sprintf("stderr_%d", i), errStr)
		}

		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("command %d failed: timeout after %v", i, timeout)
			}
			return fmt.Errorf("command %d failed (%v): %s", i, err, strings.TrimSpace(errStr))
		}
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

			expr, err := jp.ParseString(path)
			if err != nil {
				return fmt.Errorf("invalid jsonpath %s: %w", path, err)
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

	return nil
}
