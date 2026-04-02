package validate

import (
	"fmt"
	"strings"
	"time"

	"github.com/replay/replay/internal/workflow"
)

type ValidationError struct {
	Path    string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Path, e.Message)
}

type Errors []ValidationError

func (e Errors) Error() string {
	parts := make([]string, 0, len(e))
	for _, item := range e {
		parts = append(parts, item.Error())
	}
	return strings.Join(parts, "\n")
}

func Workflow(wf workflow.Workflow) error {
	var errs Errors

	if strings.TrimSpace(wf.Name) == "" {
		errs = append(errs, ValidationError{Path: "name", Message: "is required"})
	}

	if len(wf.Steps) == 0 {
		errs = append(errs, ValidationError{Path: "steps", Message: "must contain at least one step"})
	}

	seen := map[string]int{}
	for i, step := range wf.Steps {
		stepPath := fmt.Sprintf("steps[%d]", i)
		stepName := strings.TrimSpace(step.Name)
		if stepName == "" {
			errs = append(errs, ValidationError{Path: stepPath + ".name", Message: "is required"})
		} else {
			if first, ok := seen[stepName]; ok {
				errs = append(errs, ValidationError{Path: stepPath + ".name", Message: fmt.Sprintf("must be unique (already used at steps[%d])", first)})
			} else {
				seen[stepName] = i
			}
		}

		switch step.Type {
		case workflow.StepTypeHTTP:
			err := validateHTTPStep(stepPath, step)
			errs = append(errs, err...)
		case workflow.StepTypeDB:
			err := validateDBStep(stepPath, step)
			errs = append(errs, err...)
		case workflow.StepTypeShell:
			err := validateShellStep(stepPath, step)
			errs = append(errs, err...)
		case workflow.StepTypePrint:
			err := validatePrintStep(stepPath, step)
			errs = append(errs, err...)
		default:
			errs = append(errs, ValidationError{Path: stepPath + ".type", Message: "must be one of: http, db, shell, print"})
		}

		err := validateExtract(stepPath, step.Extract)
		errs = append(errs, err...)

		err = validateAssert(stepPath, step.Assert)
		errs = append(errs, err...)
	}

	if len(errs) > 0 {
		return errs
	}

	return nil
}

func validateHTTPStep(stepPath string, step workflow.Step) Errors {
	var errs Errors

	if step.Request == nil {
		errs = append(errs, ValidationError{Path: stepPath + ".request", Message: "is required for http step"})
		return errs
	}

	if strings.TrimSpace(step.Request.Method) == "" {
		errs = append(errs, ValidationError{Path: stepPath + ".request.method", Message: "is required"})
	}

	if strings.TrimSpace(step.Request.URL) == "" {
		errs = append(errs, ValidationError{Path: stepPath + ".request.url", Message: "is required"})
	}

	if step.DB != nil {
		errs = append(errs, ValidationError{Path: stepPath + ".db", Message: "must be empty for http step"})
	}

	return errs
}

func validateDBStep(stepPath string, step workflow.Step) Errors {
	var errs Errors

	db := step.DB
	if db == nil {
		// Use shortcuts if db block is missing
		var command []string
		switch c := step.Command.(type) {
		case string:
			command = []string{c}
		case []string:
			command = c
		case []any:
			for _, v := range c {
				command = append(command, fmt.Sprintf("%v", v))
			}
		}

		db = &workflow.DBRequest{
			Engine:  workflow.DBEngine(step.Engine),
			Query:   step.Query,
			Command: command,
		}
	}

	if db.Engine == "" {
		// Use postgres as default for validation if no engine is provided
		db.Engine = workflow.DBEnginePostgres
	}

	switch db.Engine {
	case workflow.DBEnginePostgres, workflow.DBEngineRedis:
	case "": // Handled above
	default:
		errs = append(errs, ValidationError{Path: stepPath + ".engine", Message: "must be one of: postgres, redis"})
	}

	query := strings.TrimSpace(db.Query)
	hasCommand := len(db.Command) > 0

	if query == "" && !hasCommand {
		errs = append(errs, ValidationError{Path: stepPath, Message: "must include query or command"})
	}

	if query != "" && hasCommand {
		errs = append(errs, ValidationError{Path: stepPath, Message: "query and command are mutually exclusive"})
	}

	for i, arg := range db.Command {
		if strings.TrimSpace(arg) == "" {
			errs = append(errs, ValidationError{Path: fmt.Sprintf("%s.command[%d]", stepPath, i), Message: "must not be empty"})
		}
	}

	if step.Request != nil {
		errs = append(errs, ValidationError{Path: stepPath + ".request", Message: "must be empty for db step"})
	}

	return errs
}

func validateShellStep(stepPath string, step workflow.Step) Errors {
	var errs Errors

	shell := step.Shell
	if shell == nil {
		var commands []string
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

		shell = &workflow.ShellRequest{
			Commands: commands,
			Dir:      step.Dir,
			Timeout:  step.Timeout,
		}
	}

	hasCommand := strings.TrimSpace(shell.Command) != ""
	hasCommands := len(shell.Commands) > 0

	if !hasCommand && !hasCommands {
		errs = append(errs, ValidationError{Path: stepPath, Message: "either command or commands is required"})
	}

	if hasCommand && hasCommands {
		errs = append(errs, ValidationError{Path: stepPath, Message: "command and commands are mutually exclusive"})
	}

	if shell.Timeout != "" {
		if _, err := time.ParseDuration(shell.Timeout); err != nil {
			errs = append(errs, ValidationError{Path: stepPath + ".timeout", Message: fmt.Sprintf("invalid duration: %v", err)})
		}
	}

	if step.Request != nil {
		errs = append(errs, ValidationError{Path: stepPath + ".request", Message: "must be empty for shell step"})
	}
	if step.DB != nil {
		errs = append(errs, ValidationError{Path: stepPath + ".db", Message: "must be empty for shell step"})
	}

	return errs
}

func validatePrintStep(stepPath string, step workflow.Step) Errors {
	var errs Errors

	if strings.TrimSpace(step.Message) == "" {
		errs = append(errs, ValidationError{Path: stepPath + ".message", Message: "is required for print step"})
	}

	if step.Request != nil {
		errs = append(errs, ValidationError{Path: stepPath + ".request", Message: "must be empty for print step"})
	}
	if step.DB != nil {
		errs = append(errs, ValidationError{Path: stepPath + ".db", Message: "must be empty for print step"})
	}
	if step.Shell != nil {
		errs = append(errs, ValidationError{Path: stepPath + ".shell", Message: "must be empty for print step"})
	}

	return errs
}

func validateExtract(stepPath string, extract map[string]string) Errors {
	var errs Errors
	for key, path := range extract {
		if strings.TrimSpace(key) == "" {
			errs = append(errs, ValidationError{Path: stepPath + ".extract", Message: "keys must not be empty"})
		}
		if strings.TrimSpace(path) == "" {
			errs = append(errs, ValidationError{Path: stepPath + ".extract." + key, Message: "jsonpath must not be empty"})
		}
	}
	return errs
}

func validateAssert(stepPath string, rules []workflow.AssertRule) Errors {
	var errs Errors
	for i, rule := range rules {
		rulePath := fmt.Sprintf("%s.assert[%d]", stepPath, i)
		if strings.TrimSpace(rule.Path) == "" {
			errs = append(errs, ValidationError{Path: rulePath + ".path", Message: "is required"})
		}
		if strings.TrimSpace(rule.Op) == "" {
			errs = append(errs, ValidationError{Path: rulePath + ".op", Message: "is required"})
		}
	}
	return errs
}
