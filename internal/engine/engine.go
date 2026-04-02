package engine

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/replay/replay/internal/runner"
	"github.com/replay/replay/internal/state"
	"github.com/replay/replay/internal/template"
	"github.com/replay/replay/internal/workflow"
)

type Engine struct {
	state      *state.Store
	httpRunner *runner.HTTPRunner
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
	}
}

func (e *Engine) Run(wf workflow.Workflow) error {
	vars := e.state.All()
	// Render workflow name if it contains variables
	wfName := template.Render(wf.Name, vars)
	if wfName == "" {
		wfName = "nameless workflow"
	}

	log.Printf("Starting workflow: %s", wfName)

	for _, step := range wf.Steps {
		vars = e.state.All()
		stepName := template.Render(step.Name, vars)
		if stepName == "" {
			stepName = fmt.Sprintf("step %s", step.Type)
		}

		log.Printf("Running step: %s [%s]", stepName, step.Type)

		switch step.Type {
		case workflow.StepTypeHTTP:
			_, err := e.httpRunner.Run(wf.Config.HTTP.BaseURL, step)
			if err != nil {
				return fmt.Errorf("step %q failed: %w", stepName, err)
			}
		case workflow.StepTypeShell:
			err := runner.Shell(step, e.state)
			if err != nil {
				return fmt.Errorf("step %q failed: %w", stepName, err)
			}
		case workflow.StepTypePrint:
			err := runner.Print(step, e.state)
			if err != nil {
				return fmt.Errorf("step %q failed: %w", stepName, err)
			}
		default:
			log.Printf("Warning: step type %q not yet implemented, skipping", step.Type)
		}
		vars = e.state.All()
	}

	log.Printf("Workflow %q completed successfully", wfName)
	return nil
}
