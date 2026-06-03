package engine

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/replay/replay/internal/parser"
	"github.com/replay/replay/internal/workflow"
)

func newTestEngine() *Engine {
	p := &parser.ParserWrapper{}
	return New(p)
}

// mockParser allows controlling which workflows are loaded for testing
type mockParser struct {
	workflows map[string][]workflow.Workflow
}

func (m *mockParser) LoadFromFile(path string) ([]workflow.Workflow, error) {
	if wfs, ok := m.workflows[path]; ok {
		return wfs, nil
	}
	return nil, fmt.Errorf("file not found: %s", path)
}

func TestEngineWorkflowScope(t *testing.T) {
	e := newTestEngine()

	wf := &workflow.Workflow{
		Name: "test-scope",
		Config: workflow.Config{
			Vars: map[string]any{
				"wf_var": "from_workflow",
			},
		},
		Steps: []workflow.Step{
			{
				Name:    "step1",
				Type:    workflow.StepTypePrint,
				Message: "value={{ wf_var }}",
			},
		},
	}

	if err := e.Run(wf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := e.State().Get("wf_var"); ok {
		t.Error("wf_var should be cleaned up after workflow exits")
	}
}

func TestEngineStepScope(t *testing.T) {
	e := newTestEngine()

	wf := &workflow.Workflow{
		Name: "test-step-scope",
		Config: workflow.Config{
			Vars: map[string]any{
				"wf_var": "workflow_value",
			},
		},
		Steps: []workflow.Step{
			{
				Name:    "set_step_var",
				Type:    workflow.StepTypeShell,
				Command: "echo ok",
			},
			{
				Name:    "verify_wf_var",
				Type:    workflow.StepTypePrint,
				Message: "wf_var={{ wf_var }}",
			},
		},
	}

	if err := e.Run(wf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngineCallIsolation(t *testing.T) {
	t.Log("call isolation test requires file-based test fixtures")
}

func TestEngineVariableValidation(t *testing.T) {
	e := newTestEngine()

	wf := &workflow.Workflow{
		Name: "test-validation",
		Config: workflow.Config{
			Vars: map[string]any{
				"user_id": 42,
			},
			Validate: []workflow.VarDef{
				{
					Name:     "user_id",
					Type:     "number",
					Required: true,
				},
				{
					Name:    "optional_var",
					Type:    "string",
					Default: "default_val",
				},
			},
		},
		Steps: []workflow.Step{
			{
				Name:    "check_vars",
				Type:    workflow.StepTypePrint,
				Message: "user={{ user_id }} opt={{ optional_var }}",
			},
		},
	}

	err := e.Run(wf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if e.State().ScopeDepth() != 1 {
		t.Errorf("expected scope depth 1 after workflow, got %d", e.State().ScopeDepth())
	}
}

func TestEngineValidationFailsOnMissingRequired(t *testing.T) {
	e := newTestEngine()

	wf := &workflow.Workflow{
		Name: "test-validation-fail",
		Config: workflow.Config{
			Validate: []workflow.VarDef{
				{
					Name:     "missing_var",
					Required: true,
				},
			},
		},
		Steps: []workflow.Step{
			{
				Name:    "should_not_run",
				Type:    workflow.StepTypePrint,
				Message: "this should not print",
			},
		},
	}

	err := e.Run(wf)
	if err == nil {
		t.Fatal("expected error for missing required variable")
	}
}

func TestEngineValidationFailsOnTypeMismatch(t *testing.T) {
	e := newTestEngine()

	wf := &workflow.Workflow{
		Name: "test-type-mismatch",
		Config: workflow.Config{
			Vars: map[string]any{
				"count": "not_a_number",
			},
			Validate: []workflow.VarDef{
				{
					Name: "count",
					Type: "number",
				},
			},
		},
		Steps: []workflow.Step{
			{
				Name:    "should_not_run",
				Type:    workflow.StepTypePrint,
				Message: "this should not print",
			},
		},
	}

	err := e.Run(wf)
	if err == nil {
		t.Fatal("expected error for type mismatch")
	}
}

func TestEngineValidationFailsOnPatternMismatch(t *testing.T) {
	e := newTestEngine()

	wf := &workflow.Workflow{
		Name: "test-pattern-mismatch",
		Config: workflow.Config{
			Vars: map[string]any{
				"username": "INVALID!",
			},
			Validate: []workflow.VarDef{
				{
					Name:    "username",
					Type:    "string",
					Pattern: "^[a-z]+$",
				},
			},
		},
		Steps: []workflow.Step{
			{
				Name:    "should_not_run",
				Type:    workflow.StepTypePrint,
				Message: "this should not print",
			},
		},
	}

	err := e.Run(wf)
	if err == nil {
		t.Fatal("expected error for pattern mismatch")
	}
}

func TestEngineCleanup(t *testing.T) {
	e := newTestEngine()

	wf := &workflow.Workflow{
		Name: "test-cleanup",
		Config: workflow.Config{},
		Steps: []workflow.Step{
			{
				Name:    "set_var",
				Type:    workflow.StepTypeShell,
				Command: "echo ok",
			},
			{
				Name:    "cleanup_step",
				Type:    workflow.StepTypePrint,
				Message: "before cleanup",
				Cleanup: []string{
					"nonexistent_var",
				},
			},
			{
				Name:    "after_cleanup",
				Type:    workflow.StepTypePrint,
				Message: "after cleanup",
			},
		},
	}

	if err := e.Run(wf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngineLoopCleanup(t *testing.T) {
	e := newTestEngine()

	e.State().Set("items", []any{"a", "b", "c"})

	wf := &workflow.Workflow{
		Name: "test-loop-cleanup",
		Config: workflow.Config{},
		Steps: []workflow.Step{
			{
				Name:    "loop_step",
				Type:    workflow.StepTypeLoop,
				ForEach: "items, item",
				Steps: []workflow.Step{
					{
						Name:    "print_item",
						Type:    workflow.StepTypePrint,
						Message: "item={{ item }}",
					},
				},
			},
			{
				Name:    "verify_cleanup",
				Type:    workflow.StepTypePrint,
				Message: "item after loop={{ item }}",
			},
		},
	}

	if err := e.Run(wf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngineScopeDepth(t *testing.T) {
	e := newTestEngine()

	if e.State().ScopeDepth() != 1 {
		t.Errorf("expected initial scope depth 1, got %d", e.State().ScopeDepth())
	}

	wf := &workflow.Workflow{
		Name: "test-depth",
		Config: workflow.Config{},
		Steps: []workflow.Step{
			{
				Name:    "step1",
				Type:    workflow.StepTypePrint,
				Message: "hello",
			},
		},
	}

	if err := e.Run(wf); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if e.State().ScopeDepth() != 1 {
		t.Errorf("expected scope depth 1 after workflow, got %d", e.State().ScopeDepth())
	}
}

func TestEngineReturnsPreservesSnapshot(t *testing.T) {
	e := newTestEngine()

	e.State().Set("pre_call_var", "before_call")

	snapshot := e.State().Snapshot()
	e.State().Set("temp_var", "temp_value")
	e.State().Restore(snapshot)

	v, ok := e.State().Get("pre_call_var")
	if !ok || v != "before_call" {
		t.Errorf("pre_call_var should be preserved, got %v", v)
	}

	if _, ok := e.State().Get("temp_var"); ok {
		t.Error("temp_var should not exist after restore")
	}
}

func TestEngineWithReturns(t *testing.T) {
	e := newTestEngine()

	e.State().Set("caller_var", "original")

	snapshot := e.State().Snapshot()

	e.State().Set("token", "new_token_value")
	e.State().Set("internal", "should_not_leak")

	returned := make(map[string]any)
	returns := []string{"token"}
	for _, k := range returns {
		if v, ok := e.State().Get(k); ok {
			returned[k] = v
		}
	}

	e.State().Restore(snapshot)
	for k, v := range returned {
		e.State().Set(k, v)
	}

	v, ok := e.State().Get("token")
	if !ok || v != "new_token_value" {
		t.Errorf("token should be preserved via returns, got %v", v)
	}

	if _, ok := e.State().Get("internal"); ok {
		t.Error("internal should not leak after restore")
	}

	v, ok = e.State().Get("caller_var")
	if !ok || v != "original" {
		t.Errorf("caller_var should be preserved, got %v", v)
	}
}

func TestEngineTTLVariables(t *testing.T) {
	e := newTestEngine()

	e.State().SetTTL("temp_key", "temp_value", 50*time.Millisecond)

	v, ok := e.State().Get("temp_key")
	if !ok || v != "temp_value" {
		t.Errorf("expected temp_value, got %v", v)
	}

	all := e.State().All()
	if _, ok := all["temp_key"]; !ok {
		t.Error("temp_key should be in All()")
	}

	time.Sleep(80 * time.Millisecond)

	if _, ok := e.State().Get("temp_key"); ok {
		t.Error("temp_key should have expired")
	}
}

func TestEngineValidationPattern(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		pattern string
		wantErr bool
	}{
		{"valid lowercase", "abc", "^[a-z]+$", false},
		{"invalid uppercase", "ABC", "^[a-z]+$", true},
		{"valid email", "user@example.com", `^[a-z]+@[a-z]+\.[a-z]+$`, false},
		{"invalid email", "not-an-email", `^[a-z]+@[a-z]+\.[a-z]+$`, true},
		{"valid alphanum", "abc123", "^[a-z0-9]+$", false},
		{"invalid alphanum", "abc-123", "^[a-z0-9]+$", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wf := &workflow.Workflow{
				Name: "test-pattern",
				Config: workflow.Config{
					Vars: map[string]any{
						"field": tt.value,
					},
					Validate: []workflow.VarDef{
						{
							Name:    "field",
							Pattern: tt.pattern,
						},
					},
				},
				Steps: []workflow.Step{
					{
						Name:    "step",
						Type:    workflow.StepTypePrint,
						Message: "{{ field }}",
					},
				},
			}

			e := newTestEngine()
			err := e.Run(wf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEngineValidationTypes(t *testing.T) {
	tests := []struct {
		name    string
		value   any
		typ     string
		wantErr bool
	}{
		{"valid string", "hello", "string", false},
		{"invalid string", 42, "string", true},
		{"valid number int", 42, "number", false},
		{"valid number float", 3.14, "number", false},
		{"valid number string", "42", "number", false},
		{"invalid number string", "abc", "number", true},
		{"valid bool", true, "bool", false},
		{"invalid bool", "true", "bool", true},
		{"valid array", []any{"a"}, "array", false},
		{"invalid array", "not_array", "array", true},
		{"valid object", map[string]any{"a": 1}, "object", false},
		{"invalid object", "not_object", "object", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateType(tt.value, tt.typ)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateType() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestEngineDirectRecursion(t *testing.T) {
	mp := &mockParser{
		workflows: map[string][]workflow.Workflow{
			"a.yaml": {
				{
					Name: "workflow-a",
					Steps: []workflow.Step{
						{
							Name: "print-a",
							Type: workflow.StepTypePrint,
							Message: "in a",
						},
						{
							Name: "call-a",
							Type: workflow.StepTypeCall,
							File: "a.yaml",
						},
					},
				},
			},
		},
	}

	e := New(mp)
	wf := &workflow.Workflow{
		Name: "test-direct-recursion",
		Steps: []workflow.Step{
			{
				Name: "call-a",
				Type: workflow.StepTypeCall,
				File: "a.yaml",
			},
		},
	}

	err := e.Run(wf)
	if err == nil {
		t.Fatal("expected error for direct recursion")
	}
	if !strings.Contains(err.Error(), "cycle detected") {
		t.Errorf("expected cycle detected error, got: %v", err)
	}
	t.Logf("error: %v", err)
}

func TestEngineMutualRecursion(t *testing.T) {
	mp := &mockParser{
		workflows: map[string][]workflow.Workflow{
			"a.yaml": {
				{
					Name: "workflow-a",
					Steps: []workflow.Step{
						{
							Name: "print-a",
							Type: workflow.StepTypePrint,
							Message: "in a",
						},
						{
							Name: "call-b",
							Type: workflow.StepTypeCall,
							File: "b.yaml",
						},
					},
				},
			},
			"b.yaml": {
				{
					Name: "workflow-b",
					Steps: []workflow.Step{
						{
							Name: "print-b",
							Type: workflow.StepTypePrint,
							Message: "in b",
						},
						{
							Name: "call-a",
							Type: workflow.StepTypeCall,
							File: "a.yaml",
						},
					},
				},
			},
		},
	}

	e := New(mp)
	wf := &workflow.Workflow{
		Name: "test-mutual-recursion",
		Steps: []workflow.Step{
			{
				Name: "call-a",
				Type: workflow.StepTypeCall,
				File: "a.yaml",
			},
		},
	}

	err := e.Run(wf)
	if err == nil {
		t.Fatal("expected error for mutual recursion")
	}
	if !strings.Contains(err.Error(), "cycle detected") {
		t.Errorf("expected cycle detected error, got: %v", err)
	}
	t.Logf("error: %v", err)
}

func TestEngineDiamondPattern(t *testing.T) {
	mp := &mockParser{
		workflows: map[string][]workflow.Workflow{
			"a.yaml": {
				{
					Name: "workflow-a",
					Steps: []workflow.Step{
						{
							Name: "print-a",
							Type: workflow.StepTypePrint,
							Message: "in a",
						},
					},
				},
			},
			"b.yaml": {
				{
					Name: "workflow-b",
					Steps: []workflow.Step{
						{
							Name: "print-b",
							Type: workflow.StepTypePrint,
							Message: "in b",
						},
						{
							Name: "call-d",
							Type: workflow.StepTypeCall,
							File: "d.yaml",
						},
					},
				},
			},
			"c.yaml": {
				{
					Name: "workflow-c",
					Steps: []workflow.Step{
						{
							Name: "print-c",
							Type: workflow.StepTypePrint,
							Message: "in c",
						},
						{
							Name: "call-d",
							Type: workflow.StepTypeCall,
							File: "d.yaml",
						},
					},
				},
			},
			"d.yaml": {
				{
					Name: "workflow-d",
					Steps: []workflow.Step{
						{
							Name: "print-d",
							Type: workflow.StepTypePrint,
							Message: "in d",
						},
					},
				},
			},
		},
	}

	e := New(mp)
	wf := &workflow.Workflow{
		Name: "test-diamond",
		Steps: []workflow.Step{
			{
				Name: "call-b",
				Type: workflow.StepTypeCall,
				File: "b.yaml",
			},
			{
				Name: "call-c",
				Type: workflow.StepTypeCall,
				File: "c.yaml",
			},
		},
	}

	err := e.Run(wf)
	if err != nil {
		t.Fatalf("unexpected error for diamond pattern: %v", err)
	}
}

func TestEngineDepthLimit(t *testing.T) {
	mp := &mockParser{
		workflows: map[string][]workflow.Workflow{
			"a.yaml": {
				{
					Name: "workflow-a",
					Steps: []workflow.Step{
						{
							Name: "call-b",
							Type: workflow.StepTypeCall,
							File: "b.yaml",
						},
					},
				},
			},
			"b.yaml": {
				{
					Name: "workflow-b",
					Steps: []workflow.Step{
						{
							Name: "call-c",
							Type: workflow.StepTypeCall,
							File: "c.yaml",
						},
					},
				},
			},
			"c.yaml": {
				{
					Name: "workflow-c",
					Steps: []workflow.Step{
						{
							Name: "call-d",
							Type: workflow.StepTypeCall,
							File: "d.yaml",
						},
					},
				},
			},
			"d.yaml": {
				{
					Name: "workflow-d",
					Steps: []workflow.Step{
						{
							Name: "print-d",
							Type: workflow.StepTypePrint,
							Message: "in d",
						},
					},
				},
			},
		},
	}

	e := New(mp)
	e.SetMaxDepth(3)

	wf := &workflow.Workflow{
		Name: "test-depth-limit",
		Steps: []workflow.Step{
			{
				Name: "call-a",
				Type: workflow.StepTypeCall,
				File: "a.yaml",
			},
		},
	}

	err := e.Run(wf)
	if err == nil {
		t.Fatal("expected error for depth limit")
	}
	if !strings.Contains(err.Error(), "call depth exceeded") {
		t.Errorf("expected call depth exceeded error, got: %v", err)
	}
	t.Logf("error: %v", err)
}

func TestEngineNormalCall(t *testing.T) {
	mp := &mockParser{
		workflows: map[string][]workflow.Workflow{
			"child.yaml": {
				{
					Name: "workflow-child",
					Steps: []workflow.Step{
						{
							Name: "print-child",
							Type: workflow.StepTypePrint,
							Message: "in child",
						},
					},
				},
			},
		},
	}

	e := New(mp)
	wf := &workflow.Workflow{
		Name: "test-normal-call",
		Steps: []workflow.Step{
			{
				Name: "print-parent",
				Type: workflow.StepTypePrint,
				Message: "in parent",
			},
			{
				Name: "call-child",
				Type: workflow.StepTypeCall,
				File: "child.yaml",
			},
		},
	}

	err := e.Run(wf)
	if err != nil {
		t.Fatalf("unexpected error for normal call: %v", err)
	}
}

func TestEngineCallSameTargetDifferentCallers(t *testing.T) {
	mp := &mockParser{
		workflows: map[string][]workflow.Workflow{
			"a.yaml": {
				{
					Name: "workflow-a",
					Steps: []workflow.Step{
						{
							Name: "call-b-target",
							Type: workflow.StepTypeCall,
							File: "b.yaml",
							Target: "step1",
						},
					},
				},
			},
			"b.yaml": {
				{
					Name: "workflow-b",
					Steps: []workflow.Step{
						{
							Name: "step1",
							Type: workflow.StepTypePrint,
							Message: "step1 in b",
						},
						{
							Name: "step2",
							Type: workflow.StepTypePrint,
							Message: "step2 in b",
						},
					},
				},
			},
			"c.yaml": {
				{
					Name: "workflow-c",
					Steps: []workflow.Step{
						{
							Name: "call-b-target",
							Type: workflow.StepTypeCall,
							File: "b.yaml",
							Target: "step1",
						},
					},
				},
			},
		},
	}

	e := New(mp)
	wf := &workflow.Workflow{
		Name: "test-same-target-different-callers",
		Steps: []workflow.Step{
			{
				Name: "call-a",
				Type: workflow.StepTypeCall,
				File: "a.yaml",
			},
			{
				Name: "call-c",
				Type: workflow.StepTypeCall,
				File: "c.yaml",
			},
		},
	}

	err := e.Run(wf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEngineCallStackResetBetweenWorkflows(t *testing.T) {
	mp := &mockParser{
		workflows: map[string][]workflow.Workflow{
			"a.yaml": {
				{
					Name: "workflow-a",
					Steps: []workflow.Step{
						{
							Name: "print-a",
							Type: workflow.StepTypePrint,
							Message: "in a",
						},
					},
				},
			},
		},
	}

	e := New(mp)

	// First workflow
	wf1 := &workflow.Workflow{
		Name: "workflow-1",
		Steps: []workflow.Step{
			{
				Name: "call-a",
				Type: workflow.StepTypeCall,
				File: "a.yaml",
			},
		},
	}
	if err := e.Run(wf1); err != nil {
		t.Fatalf("first workflow failed: %v", err)
	}

	// Second workflow should work fine (call stack reset)
	wf2 := &workflow.Workflow{
		Name: "workflow-2",
		Steps: []workflow.Step{
			{
				Name: "call-a",
				Type: workflow.StepTypeCall,
				File: "a.yaml",
			},
		},
	}
	if err := e.Run(wf2); err != nil {
		t.Fatalf("second workflow failed: %v", err)
	}
}
