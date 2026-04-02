package runner

import (
	"fmt"

	"github.com/replay/replay/internal/state"
	"github.com/replay/replay/internal/template"
	"github.com/replay/replay/internal/workflow"
)

// Print renders and displays a message to stdout.
func Print(step workflow.Step, store *state.Store) error {
	if step.Message == "" {
		return fmt.Errorf("message is empty")
	}

	msg := template.Render(step.Message, store.All())

	fmt.Printf(" [PRINT] %s\n", msg)
	return nil
}
