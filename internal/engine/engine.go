package engine

import (
	"fmt"
	"os"
	"strings"

	"time"

	"github.com/replay/replay/internal/reporter"
	"github.com/replay/replay/internal/runner"
	"github.com/replay/replay/internal/state"
	"github.com/replay/replay/internal/template"
	"github.com/replay/replay/internal/workflow"
)

type Parser interface {
	LoadFromFile(path string) ([]workflow.Workflow, error)
}

type Engine struct {
	state      *state.Store
	httpRunner *runner.HTTPRunner
	reporter   *reporter.Reporter
	parser     Parser
}

func New(p Parser) *Engine {
	s := state.NewStore()
	rep := reporter.New()
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			s.Set(pair[0], pair[1])
		}
	}
	return &Engine{
		state:      s,
		httpRunner: runner.NewHTTPRunner(s, rep),
		reporter:   rep,
		parser:     p,
	}
}

func (e *Engine) Run(wf *workflow.Workflow) error {
	// Set global config into state for use in templates (e.g., baseURL)
	e.state.Set("config", wf.Config)

	vars := e.state.All()
	// Ensure the workflow name itself is in the state for interpolation
	e.state.Set("name", wf.Name)

	// Render workflow name if it contains variables
	wfName := template.Render(wf.Name, vars)
	if wfName == "" {
		wfName = "nameless workflow"
	}

	wfStart := time.Now()
	e.reporter.WorkflowStarted(wfName)

	if err := e.ExecuteSteps(wf.Steps, wf.Config); err != nil {
		e.reporter.WorkflowFinished(wfName, false, time.Since(wfStart))
		return err
	}

	e.reporter.WorkflowFinished(wfName, true, time.Since(wfStart))
	return nil
}

func (e *Engine) ExecuteSteps(steps []workflow.Step, config workflow.Config) error {
	for _, step := range steps {
		vars := e.state.All()
		stepName := template.Render(step.Name, vars)
		if stepName == "" {
			stepName = fmt.Sprintf("step %s", step.Type)
		}

		e.reporter.StepStarted(stepName, string(step.Type))
		stepStart := time.Now()

		var err error
		switch step.Type {
		case workflow.StepTypeHTTP:
			_, err = e.httpRunner.Run(config.HTTP, step)
		case workflow.StepTypeShell:
			err = runner.Shell(step, e.state)
		case workflow.StepTypeDB:
			err = runner.DB(config, step, e.state)
		case workflow.StepTypePrint:
			err = runner.Print(step, e.state)
		case workflow.StepTypeLoop:
			err = e.ExecuteLoop(step, config)
		case workflow.StepTypeCall:
			err = e.ExecuteCall(step, config)
		case workflow.StepTypeIf:
			err = e.ExecuteIf(step, config)
		default:
			err = fmt.Errorf("step type %q not yet implemented", step.Type)
		}

		duration := time.Since(stepStart)
		if err != nil {
			if step.IgnoreError {
				e.reporter.StepFailed(err, duration, true)
			} else {
				e.reporter.StepFailed(err, duration, false)
				return fmt.Errorf("step %q failed: %w", stepName, err)
			}
		} else {
			e.reporter.StepPassed(duration)
		}
	}
	return nil
}

func (e *Engine) ExecuteLoop(step workflow.Step, config workflow.Config) error {
	if step.ForEach == "" {
		return fmt.Errorf("loop step requires 'foreach' field")
	}

	// format: "list, item"
	parts := strings.Split(step.ForEach, ",")
	if len(parts) != 2 {
		return fmt.Errorf("foreach must be in format 'list, item'")
	}
	listKey := strings.TrimSpace(parts[0])
	itemKey := strings.TrimSpace(parts[1])

	val, ok := e.state.Get(listKey)
	if !ok {
		return fmt.Errorf("variable %q not found for loop", listKey)
	}

	// Handle both []any and []map[string]any etc.
	// Since we use json.Unmarshal into any, it should be []any for arrays
	items, ok := val.([]any)
	if !ok {
		return fmt.Errorf("variable %q is not a list", listKey)
	}

	for i, item := range items {
		// Set loop context
		e.state.Set(itemKey, item)
		e.state.Set("index", i)

		// Execute nested steps
		if err := e.ExecuteSteps(step.Steps, config); err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) ExecuteCall(step workflow.Step, config workflow.Config) error {
	vars := e.state.All()
	filePath := template.Render(step.File, vars)
	target := template.Render(step.Target, vars)

	if filePath == "" {
		return fmt.Errorf("call step requires 'file' field")
	}

	wfs, err := e.parser.LoadFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load called file %q: %w", filePath, err)
	}

	// Apply input variables from 'with' block
	for k, v := range step.With {
		// Interpolate the value before setting it
		if s, ok := v.(string); ok {
			e.state.Set(k, template.Render(s, vars))
		} else {
			e.state.Set(k, v)
		}
	}

	// Case 1: Call a specific step within a file
	if target != "" {
		for _, wf := range wfs {
			for _, s := range wf.Steps {
				if s.Name == target {
					// Found the target step, execute ONLY this step
					return e.ExecuteSteps([]workflow.Step{s}, wf.Config)
				}
			}
		}
		return fmt.Errorf("step %q not found in file %q", target, filePath)
	}

	// Case 2: Call the whole workflow(s)
	for _, wf := range wfs {
		if err := e.ExecuteSteps(wf.Steps, wf.Config); err != nil {
			return err
		}
	}

	return nil
}

func (e *Engine) ExecuteIf(step workflow.Step, config workflow.Config) error {
	vars := e.state.All()
	if len(step.Condition) != 3 {
		return fmt.Errorf("if step condition must be in format [path, op, value]")
	}

	ae := runner.NewAssertionEngine(vars)
	rule := workflow.AssertRule{
		Path:  step.Condition[0],
		Op:    step.Condition[1],
		Value: step.Condition[2],
	}

	// Use empty data since assertions are against state path
	err := ae.Check(rule, nil)
	if err == nil {
		// Condition met (then)
		return e.ExecuteSteps(step.Then, config)
	} else {
		// Condition failed (else)
		return e.ExecuteSteps(step.Else, config)
	}
}
