package engine

import (
	"fmt"
	"log"

	"github.com/replay/replay/internal/runner"
	"github.com/replay/replay/internal/state"
	"github.com/replay/replay/internal/workflow"
)

type Engine struct {
	state      *state.Store
	httpRunner *runner.HTTPRunner
}

func New() *Engine {
	s := state.NewStore()
	return &Engine{
		state:      s,
		httpRunner: runner.NewHTTPRunner(s),
	}
}

func (e *Engine) Run(wf workflow.Workflow) error {
	log.Printf("Starting workflow: %s", wf.Name)

	for _, step := range wf.Steps {
		log.Printf("Running step: %s [%s]", step.Name, step.Type)

		switch step.Type {
		case workflow.StepTypeHTTP:
			_, err := e.httpRunner.Run(wf.Config.HTTP.BaseURL, step)
			if err != nil {
				return fmt.Errorf("step %q failed: %w", step.Name, err)
			}
		default:
			log.Printf("Warning: step type %q not yet implemented, skipping", step.Type)
		}
	}

	log.Printf("Workflow %q completed successfully", wf.Name)
	return nil
}
