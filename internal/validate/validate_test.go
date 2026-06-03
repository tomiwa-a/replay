package validate

import (
	"strings"
	"testing"

	"github.com/replay/replay/internal/workflow"
)

func TestWorkflow_ValidHTTPAndDB(t *testing.T) {
	wf := workflow.Workflow{
		Name: "ok",
		Steps: []workflow.Step{
			{
				Name: "login",
				Type: workflow.StepTypeHTTP,
				Request: &workflow.HTTPRequest{
					Method: "POST",
					URL:    "/auth/login",
				},
				Extract: map[string]string{"token": "$.data.token"},
				Assert:  []workflow.AssertRule{{Path: "$.status", Op: "eq", Value: 200}},
			},
			{
				Name: "check-db",
				Type: workflow.StepTypeDB,
				DB: &workflow.DBRequest{
					Engine: workflow.DBEnginePostgres,
					Query:  "select 1",
				},
			},
		},
	}

	if err := Workflow(wf); err != nil {
		t.Fatalf("expected no validation errors, got %v", err)
	}
}

func TestWorkflow_AggregatesValidationErrors(t *testing.T) {
	wf := workflow.Workflow{
		Steps: []workflow.Step{
			{
				Name: "dup",
				Type: workflow.StepTypeHTTP,
			},
			{
				Name: "dup",
				Type: "invalid",
			},
		},
	}

	err := Workflow(wf)
	if err == nil {
		t.Fatal("expected validation errors, got nil")
	}

	msg := err.Error()
	requiredChecks := []string{
		"name: is required",
		"steps[0].request: is required for http step",
		"steps[1].name: must be unique",
		"steps[1].type: must be one of: http, db",
	}

	for _, check := range requiredChecks {
		if !strings.Contains(msg, check) {
			t.Fatalf("expected error to contain %q, got %q", check, msg)
		}
	}
}

func TestWorkflow_DBRules(t *testing.T) {
	wf := workflow.Workflow{
		Name: "db-rules",
		Steps: []workflow.Step{
			{
				Name: "s1",
				Type: workflow.StepTypeDB,
				DB: &workflow.DBRequest{
					Engine:  workflow.DBEngineRedis,
					Query:   "GET x",
					Command: []string{"GET", "x"},
				},
			},
		},
	}

	err := Workflow(wf)
	if err == nil {
		t.Fatal("expected db validation error, got nil")
	}

	if !strings.Contains(err.Error(), "query and command are mutually exclusive") {
		t.Fatalf("expected mutual exclusivity error, got %v", err)
	}
}

func TestDetectCycles_NoCalls(t *testing.T) {
	wf := workflow.Workflow{
		Name: "no-calls",
		Steps: []workflow.Step{
			{Name: "step1", Type: workflow.StepTypePrint, Message: "hello"},
		},
	}

	if err := DetectCycles(wf, "no-calls.yaml"); err != nil {
		t.Fatalf("expected no cycles, got %v", err)
	}
}

func TestDetectCycles_NoCycle(t *testing.T) {
	wf := workflow.Workflow{
		Name: "no-cycle",
		Steps: []workflow.Step{
			{Name: "call-a", Type: workflow.StepTypeCall, File: "a.yaml"},
			{Name: "call-b", Type: workflow.StepTypeCall, File: "b.yaml"},
		},
	}

	if err := DetectCycles(wf, "main.yaml"); err != nil {
		t.Fatalf("expected no cycles, got %v", err)
	}
}

func TestDetectCycles_DirectCycle(t *testing.T) {
	wf := workflow.Workflow{
		Name: "a.yaml",
		Steps: []workflow.Step{
			{Name: "call-a", Type: workflow.StepTypeCall, File: "a.yaml"},
		},
	}

	err := DetectCycles(wf, "a.yaml")
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
	if !strings.Contains(err.Error(), "cycle detected") {
		t.Errorf("expected cycle detected error, got: %v", err)
	}
}

func TestDetectCycles_MutualCycle(t *testing.T) {
	wf := workflow.Workflow{
		Name: "a.yaml",
		Steps: []workflow.Step{
			{Name: "call-b", Type: workflow.StepTypeCall, File: "b.yaml"},
		},
	}

	// Cross-file cycles cannot be detected statically without loading the called file
	if err := DetectCycles(wf, "a.yaml"); err != nil {
		t.Fatalf("expected no cycles (cross-file cycles require runtime detection), got %v", err)
	}
}

func TestDetectCycles_SelfCallWithTarget(t *testing.T) {
	wf := workflow.Workflow{
		Name: "a.yaml",
		Steps: []workflow.Step{
			{Name: "call-a-login", Type: workflow.StepTypeCall, File: "a.yaml", Target: "login"},
			{Name: "call-a-login-again", Type: workflow.StepTypeCall, File: "a.yaml", Target: "login"},
		},
	}

	// Calling the same target twice is not a cycle (they're sequential calls)
	err := DetectCycles(wf, "a.yaml")
	if err != nil {
		t.Fatalf("expected no cycle (sequential calls are not cycles), got: %v", err)
	}
}

func TestDetectCycles_SameFileDifferentTargets(t *testing.T) {
	wf := workflow.Workflow{
		Name: "a.yaml",
		Steps: []workflow.Step{
			{Name: "call-a-login", Type: workflow.StepTypeCall, File: "a.yaml", Target: "login"},
			{Name: "call-a-logout", Type: workflow.StepTypeCall, File: "a.yaml", Target: "logout"},
		},
	}

	if err := DetectCycles(wf, "a.yaml"); err != nil {
		t.Fatalf("expected no cycles (different targets), got %v", err)
	}
}

func TestDetectCycles_NestedIfCycle(t *testing.T) {
	wf := workflow.Workflow{
		Name: "a.yaml",
		Steps: []workflow.Step{
			{
				Name:      "check",
				Type:      workflow.StepTypeIf,
				Condition: []string{"status", "==", "200"},
				Then: []workflow.Step{
					{Name: "call-a", Type: workflow.StepTypeCall, File: "a.yaml"},
				},
			},
		},
	}

	err := DetectCycles(wf, "a.yaml")
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
}

func TestDetectCycles_NestedLoopCycle(t *testing.T) {
	wf := workflow.Workflow{
		Name: "a.yaml",
		Steps: []workflow.Step{
			{
				Name:      "loop",
				Type:      workflow.StepTypeLoop,
				ForEach: "items, item",
				Steps: []workflow.Step{
					{Name: "call-a", Type: workflow.StepTypeCall, File: "a.yaml"},
				},
			},
		},
	}

	err := DetectCycles(wf, "a.yaml")
	if err == nil {
		t.Fatal("expected cycle error, got nil")
	}
	if !strings.Contains(err.Error(), "cycle detected") {
		t.Errorf("expected cycle detected error, got: %v", err)
	}
}
