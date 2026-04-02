package runner

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/replay/replay/internal/state"
	"github.com/replay/replay/internal/template"
	"github.com/replay/replay/internal/workflow"

	"github.com/ohler55/ojg/jp"
)

// Shell executes a shell command.
func Shell(step workflow.Step, store *state.Store) error {
	if step.Shell == nil {
		return fmt.Errorf("shell configuration is missing")
	}

	// Render command and working directory
	cmdStr := template.Render(step.Shell.Command, store.All())

	var dir string
	if step.Shell.Dir != "" {
		dir = template.Render(step.Shell.Dir, store.All())
	}

	// Prepare the command
	// Note: We use "sh -c" on Unix/macOS for flexibility
	cmd := exec.Command("sh", "-c", cmdStr)
	if dir != "" {
		cmd.Dir = dir
	}

	// Capture stdout/stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run the command
	err := cmd.Run()

	output := stdout.String()
	store.Set("stdout", output)
	store.Set("stderr", stderr.String())

	if len(step.Extract) > 0 {
		var data interface{}
		errJSON := json.Unmarshal([]byte(output), &data)

		for varName, path := range step.Extract {
			if errJSON != nil {
				if path == "$" {
					store.Set(varName, output)
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

	if err != nil {
		return fmt.Errorf("shell command failed (%v): %s", err, strings.TrimSpace(stderr.String()))
	}

	return nil
}
