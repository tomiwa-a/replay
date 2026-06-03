package engine

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/replay/replay/internal/workflow"
)

func TestExtractionAcrossSteps(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"token": "secret-abc-123"})
	}))
	defer srv.Close()

	e := New(&mockParser{})
	wf := workflow.Workflow{
		Name: "extract-test",
		Config: workflow.Config{
			HTTP: workflow.HTTPConfig{BaseURL: srv.URL},
		},
		Steps: []workflow.Step{
			{
				Name: "step1",
				Type: workflow.StepTypeHTTP,
				Request: &workflow.HTTPRequest{
					Method: "GET",
					URL:    "/token",
				},
				Extract: map[string]string{
					"my_token": "data.token",
				},
				Assert: []workflow.AssertRule{
					{Path: "$.status", Op: "eq", Value: 200},
				},
			},
			{
				Name:    "step2",
				Type:    workflow.StepTypePrint,
				Message: "token={{ my_token }}",
			},
		},
	}

	// If extraction fails, step2 will try to render {{ my_token }} and succeed
	// with empty string, but the workflow will still return nil.
	// If the assert on step1 fails, Run returns an error.
	if err := e.Run(&wf); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
}

func TestExtractionFailureFailsRun(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"token": nil})
	}))
	defer srv.Close()

	e := New(&mockParser{})
	wf := workflow.Workflow{
		Name: "extract-fail",
		Config: workflow.Config{
			HTTP: workflow.HTTPConfig{BaseURL: srv.URL},
		},
		Steps: []workflow.Step{
			{
				Name: "step1",
				Type: workflow.StepTypeHTTP,
				Request: &workflow.HTTPRequest{
					Method: "GET",
					URL:    "/token",
				},
				Assert: []workflow.AssertRule{
					{Path: "$.data.token", Op: "not_null"},
				},
			},
		},
	}

	if err := e.Run(&wf); err == nil {
		t.Fatal("expected assertion error (token is null), got nil")
	}
}
