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
