package parser

import (
	"fmt"
	"path/filepath"

	"github.com/replay/replay/internal/template"
	"github.com/replay/replay/internal/workflow"
)

// ResolveIncludes loads and merges any included workflow files into the main workflow.
// Included steps are prepended to the workflow's steps in the order they appear.
// Parameters in the 'with' block are applied to the included steps via template substitution.
func ResolveIncludes(wf *workflow.Workflow) error {
	if len(wf.Include) == 0 {
		return nil
	}

	var allSteps []workflow.Step
	for _, inc := range wf.Include {
		if inc.File == "" {
			return fmt.Errorf("include entry requires 'file' field")
		}

		resolvedPath := inc.File
		if wf.BaseDir != "" && !filepath.IsAbs(resolvedPath) {
			resolvedPath = filepath.Join(wf.BaseDir, resolvedPath)
		}

		wfs, err := LoadFromFile(resolvedPath)
		if err != nil {
			return fmt.Errorf("failed to load included file %q: %w", inc.File, err)
		}

		params := make(map[string]any)
		for k, v := range inc.With {
			params[k] = v
		}

		for _, included := range wfs {
			if err := ResolveIncludes(&included); err != nil {
				return err
			}
			for i := range included.Steps {
				applyParamsToStep(&included.Steps[i], params)
			}
			allSteps = append(allSteps, included.Steps...)
		}
	}

	allSteps = append(allSteps, wf.Steps...)
	wf.Steps = allSteps
	return nil
}

// applyParamsToStep substitutes template variables in step fields with the provided parameters.
func applyParamsToStep(step *workflow.Step, params map[string]any) {
	if step.Message != "" {
		step.Message = template.Render(step.Message, params)
	}
	if step.ForEach != "" {
		step.ForEach = template.Render(step.ForEach, params)
	}
	if step.File != "" {
		step.File = template.Render(step.File, params)
	}
	if step.Target != "" {
		step.Target = template.Render(step.Target, params)
	}
	if step.Query != "" {
		step.Query = template.Render(step.Query, params)
	}
	if step.Dir != "" {
		step.Dir = template.Render(step.Dir, params)
	}
	if step.Timeout != "" {
		step.Timeout = template.Render(step.Timeout, params)
	}

	for i, cmd := range step.Commands {
		step.Commands[i] = template.Render(cmd, params)
	}

	if step.Request != nil {
		if step.Request.URL != "" {
			step.Request.URL = template.Render(step.Request.URL, params)
		}
		for k, v := range step.Request.Headers {
			step.Request.Headers[k] = template.Render(v, params)
		}
		if bodyStr, ok := step.Request.Body.(string); ok {
			step.Request.Body = template.Render(bodyStr, params)
		}
	}

	if step.Shell != nil {
		if step.Shell.Command != "" {
			step.Shell.Command = template.Render(step.Shell.Command, params)
		}
		for i, cmd := range step.Shell.Commands {
			step.Shell.Commands[i] = template.Render(cmd, params)
		}
		if step.Shell.Dir != "" {
			step.Shell.Dir = template.Render(step.Shell.Dir, params)
		}
	}

	if step.DB != nil {
		if step.DB.Query != "" {
			step.DB.Query = template.Render(step.DB.Query, params)
		}
		for i, arg := range step.DB.Command {
			step.DB.Command[i] = template.Render(arg, params)
		}
	}

	if step.With != nil {
		for k, v := range step.With {
			if strVal, ok := v.(string); ok {
				step.With[k] = template.Render(strVal, params)
			}
		}
	}

	for i, cond := range step.Condition {
		step.Condition[i] = template.Render(cond, params)
	}
}