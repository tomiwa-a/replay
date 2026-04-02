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
	var commands []string
	var timeoutStr string
	var dirRaw string

	if step.Shell != nil {
		if step.Shell.Command != "" {
			commands = append(commands, step.Shell.Command)
		} else if len(step.Shell.Commands) > 0 {
			commands = step.Shell.Commands
		}
		timeoutStr = step.Shell.Timeout
		dirRaw = step.Shell.Dir
	} else {
		// Shortcuts
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

	// Assertions
	ae := NewAssertionEngine(store.All())
	for _, rule := range step.Assert {
		if err := ae.Check(rule, finalOutput); err != nil {
			return fmt.Errorf("assertion failed: %w", err)
		}
	}

	return nil
}
