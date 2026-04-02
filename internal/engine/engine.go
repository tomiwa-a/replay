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

type Engine struct {
	state      *state.Store
	httpRunner *runner.HTTPRunner
	reporter   *reporter.Reporter
}

func New() *Engine {
	s := state.NewStore()
	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		if len(pair) == 2 {
			s.Set(pair[0], pair[1])
		}
	}
	return &Engine{
		state:      s,
		httpRunner: runner.NewHTTPRunner(s),
		reporter:   reporter.New(),
	}
}

func (e *Engine) Run(wf *workflow.Workflow) error {
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

	for _, step := range wf.Steps {
		vars = e.state.All()
		stepName := template.Render(step.Name, vars)
		if stepName == "" {
			stepName = fmt.Sprintf("step %s", step.Type)
		}

		e.reporter.StepStarted(stepName, string(step.Type))
		stepStart := time.Now()

		var err error
		switch step.Type {
		case workflow.StepTypeHTTP:
			_, err = e.httpRunner.Run(wf.Config.HTTP.BaseURL, step)
		case workflow.StepTypeShell:
			err = runner.Shell(step, e.state)
		case workflow.StepTypeDB:
			err = runner.DB(wf.Config, step, e.state)
		case workflow.StepTypePrint:
			err = runner.Print(step, e.state)
		default:
			err = fmt.Errorf("step type %q not yet implemented", step.Type)
		}

		duration := time.Since(stepStart)
		if err != nil {
			if step.IgnoreError {
				e.reporter.StepFailed(err, duration, true)
			} else {
				e.reporter.StepFailed(err, duration, false)
				e.reporter.WorkflowFinished(wfName, false, time.Since(wfStart))
				return fmt.Errorf("step %q failed: %w", stepName, err)
			}
		} else {
			e.reporter.StepPassed(duration)
		}

		vars = e.state.All()
	}

	e.reporter.WorkflowFinished(wfName, true, time.Since(wfStart))
	return nil
}
