package engine

import (
	"fmt"
	"os"
	"regexp"
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

type callEntry struct {
	File string
	Step string
}

func (e callEntry) String() string {
	if e.Step != "" {
		return fmt.Sprintf("%s:%s", e.File, e.Step)
	}
	return e.File
}

type Engine struct {
	state      *state.Store
	httpRunner *runner.HTTPRunner
	reporter   *reporter.Reporter
	parser     Parser
	debug      bool
	callStack  []callEntry
	maxDepth   int
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
		maxDepth:   100,
	}
}

func (e *Engine) SetDebug(debug bool) {
	e.debug = debug
}

func (e *Engine) IsDebug() bool {
	return e.debug
}

func (e *Engine) State() *state.Store {
	return e.state
}

func (e *Engine) SetMaxDepth(depth int) {
	if depth > 0 {
		e.maxDepth = depth
	}
}

func (e *Engine) formatStack() string {
	parts := make([]string, len(e.callStack))
	for i, entry := range e.callStack {
		parts[i] = entry.String()
	}
	return strings.Join(parts, " → ")
}

func (e *Engine) Run(wf *workflow.Workflow) error {
	e.state.Enter(state.ScopeWorkflow)

	e.state.Set("config", wf.Config)
	for k, v := range wf.Config.Vars {
		e.state.Set(k, v)
	}

	if err := e.validateWorkflowVars(wf); err != nil {
		e.state.Exit()
		return err
	}

	vars := e.state.All()
	e.state.Set("name", wf.Name)

	wfName := template.Render(wf.Name, vars)
	if wfName == "" {
		wfName = "nameless workflow"
	}

	wfStart := time.Now()
	e.reporter.WorkflowStarted(wfName)

	if err := e.ExecuteSteps(wf.Steps, wf.Config); err != nil {
		e.reporter.WorkflowFinished(wfName, false, time.Since(wfStart))
		e.state.Exit()
		return err
	}

	e.reporter.WorkflowFinished(wfName, true, time.Since(wfStart))
	e.state.Exit()
	return nil
}

func (e *Engine) ExecuteSteps(steps []workflow.Step, config workflow.Config) error {
	for _, step := range steps {
		e.state.Enter(state.ScopeStep)

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

		if len(step.Cleanup) > 0 {
			for _, k := range step.Cleanup {
				e.state.Delete(k)
			}
		}

		duration := time.Since(stepStart)
		if err != nil {
			if step.IgnoreError {
				e.reporter.StepFailed(err, duration, true)
			} else {
				e.reporter.StepFailed(err, duration, false)
				e.state.Exit()
				return fmt.Errorf("step %q failed: %w", stepName, err)
			}
		} else {
			// Promote extracted values from step scope to workflow scope
			// so subsequent steps can access them via {{ var }}.
			e.state.Promote()
			e.reporter.StepPassed(duration)
		}

		e.state.Exit()
	}
	return nil
}

func (e *Engine) ExecuteLoop(step workflow.Step, config workflow.Config) error {
	if step.ForEach == "" {
		return fmt.Errorf("loop step requires 'foreach' field")
	}

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

	items, ok := val.([]any)
	if !ok {
		return fmt.Errorf("variable %q is not a list", listKey)
	}

	for i, item := range items {
		e.state.Set(itemKey, item)
		e.state.Set("index", i)

		if err := e.ExecuteSteps(step.Steps, config); err != nil {
			return err
		}
	}

	e.state.Delete(itemKey)
	e.state.Delete("index")

	return nil
}

func (e *Engine) ExecuteCall(step workflow.Step, config workflow.Config) error {
	vars := e.state.All()
	filePath := template.Render(step.File, vars)
	target := template.Render(step.Target, vars)

	if filePath == "" {
		return fmt.Errorf("call step requires 'file' field")
	}

	entry := callEntry{File: filePath, Step: target}

	// Check for cycle
	for _, s := range e.callStack {
		if s.File == entry.File && s.Step == entry.Step {
			stack := e.formatStack()
			return fmt.Errorf("cycle detected: %s → %s", stack, entry)
		}
	}

	// Check depth limit
	if e.maxDepth > 0 && len(e.callStack) >= e.maxDepth {
		return fmt.Errorf("call depth exceeded (%d): %s", e.maxDepth, e.formatStack())
	}

	// Push onto call stack
	e.callStack = append(e.callStack, entry)
	defer func() { e.callStack = e.callStack[:len(e.callStack)-1] }()

	wfs, err := e.parser.LoadFromFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to load called file %q: %w", filePath, err)
	}

	snapshot := e.state.Snapshot()

	for k, v := range step.With {
		if s, ok := v.(string); ok {
			e.state.Set(k, template.Render(s, vars))
		} else {
			e.state.Set(k, v)
		}
	}

	executeSteps := func(steps []workflow.Step, cfg workflow.Config) error {
		return e.ExecuteSteps(steps, cfg)
	}

	if target != "" {
		for _, wf := range wfs {
			for _, s := range wf.Steps {
				if s.Name == target {
					if err := executeSteps([]workflow.Step{s}, wf.Config); err != nil {
						e.state.Restore(snapshot)
						return err
					}
					e.applyReturns(step.Returns, snapshot)
					return nil
				}
			}
		}
		e.state.Restore(snapshot)
		return fmt.Errorf("step %q not found in file %q", target, filePath)
	}

	for _, wf := range wfs {
		if err := executeSteps(wf.Steps, wf.Config); err != nil {
			e.state.Restore(snapshot)
			return err
		}
	}

	e.applyReturns(step.Returns, snapshot)
	return nil
}

func (e *Engine) applyReturns(returns []string, snapshot map[string]any) {
	if len(returns) == 0 {
		e.state.Restore(snapshot)
		return
	}

	returned := make(map[string]any)
	for _, k := range returns {
		if v, ok := e.state.Get(k); ok {
			returned[k] = v
		}
	}

	e.state.Restore(snapshot)

	for k, v := range returned {
		e.state.Set(k, v)
	}
}

func (e *Engine) ExecuteIf(step workflow.Step, config workflow.Config) error {
	vars := e.state.All()
	if len(step.Condition) != 3 {
		return fmt.Errorf("if step condition must be in format [path, op, value]")
	}

	ae := runner.NewAssertionEngine(vars)

	path := step.Condition[0]
	op := step.Condition[1]
	expected := step.Condition[2]

	rule := workflow.AssertRule{
		Path:  path,
		Op:    op,
		Value: expected,
	}

	err := ae.Check(rule, vars)
	if err == nil {
		return e.ExecuteSteps(step.Then, config)
	} else {
		return e.ExecuteSteps(step.Else, config)
	}
}

func (e *Engine) validateWorkflowVars(wf *workflow.Workflow) error {
	for _, def := range wf.Config.Validate {
		val, ok := e.state.Get(def.Name)

		if def.Required && !ok {
			return fmt.Errorf("variable %q is required but not set in workflow %q", def.Name, wf.Name)
		}

		if !ok {
			if def.Default != nil {
				e.state.Set(def.Name, def.Default)
			}
			continue
		}

		if def.Type != "" {
			if err := validateType(val, def.Type); err != nil {
				return fmt.Errorf("variable %q: %w", def.Name, err)
			}
		}

		if def.Pattern != "" {
			s := fmt.Sprintf("%v", val)
			matched, err := regexp.MatchString(def.Pattern, s)
			if err != nil {
				return fmt.Errorf("variable %q: invalid pattern %q: %w", def.Name, def.Pattern, err)
			}
			if !matched {
				return fmt.Errorf("variable %q value %q does not match pattern %q", def.Name, s, def.Pattern)
			}
		}
	}
	return nil
}

func validateType(val any, typ string) error {
	switch typ {
	case "string":
		if _, ok := val.(string); !ok {
			return fmt.Errorf("expected type string, got %T", val)
		}
	case "number":
		switch v := val.(type) {
		case int, int64, float64, float32:
			_ = v
		case string:
			_, err := fmt.Sscanf(v, "%f", new(float64))
			if err != nil {
				return fmt.Errorf("expected type number, got string %q", v)
			}
		default:
			return fmt.Errorf("expected type number, got %T", val)
		}
	case "bool":
		if _, ok := val.(bool); !ok {
			return fmt.Errorf("expected type bool, got %T", val)
		}
	case "array":
		switch val.(type) {
		case []any, []string, []int, []float64:
		default:
			return fmt.Errorf("expected type array, got %T", val)
		}
	case "object":
		if _, ok := val.(map[string]any); !ok {
			return fmt.Errorf("expected type object, got %T", val)
		}
	}
	return nil
}
