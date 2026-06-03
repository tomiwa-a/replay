package runner

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/replay/replay/internal/reporter"
	"github.com/replay/replay/internal/state"
	"github.com/replay/replay/internal/workflow"
)

func TestHTTPRunner_Run(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id":    123,
			"token": "secret-token",
		})
	}))
	defer s.Close()

	store := state.NewStore()
	rep := reporter.New()
	runner := NewHTTPRunner(store, rep)

	step := workflow.Step{
		Name: "login",
		Type: workflow.StepTypeHTTP,
		Request: &workflow.HTTPRequest{
			Method: "GET",
			URL:    "/auth",
		},
		Extract: map[string]string{
			"my_token": "$.data.token",
		},
	}

	config := workflow.HTTPConfig{BaseURL: s.URL}
	_, err := runner.Run(config, step)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	token, ok := store.Get("my_token")
	if !ok || token != "secret-token" {
		t.Errorf("expected secret-token extracted, got %v", token)
	}
}

func TestHTTPRunner_AssertionFailure(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]any{
			"status": "error",
			"code":   "INTERNAL_ERROR",
		})
	}))
	defer s.Close()

	store := state.NewStore()
	rep := reporter.New()
	runner := NewHTTPRunner(store, rep)

	step := workflow.Step{
		Name: "failing-step",
		Type: workflow.StepTypeHTTP,
		Request: &workflow.HTTPRequest{
			Method: "GET",
			URL:    "/fail",
		},
		Assert: []workflow.AssertRule{
			{Path: "$.status", Op: "eq", Value: 200},
		},
	}

	config := workflow.HTTPConfig{BaseURL: s.URL}
	_, err := runner.Run(config, step)
	if err == nil {
		t.Fatal("expected assertion error, got nil")
	}
	if !strings.Contains(err.Error(), "assertion failed") {
		t.Errorf("expected assertion failed error, got: %v", err)
	}
	t.Logf("got expected error: %v", err)
}

func TestHTTPRunner_AssertionPass(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"email": "test@example.com"})
	}))
	defer s.Close()

	store := state.NewStore()
	rep := reporter.New()
	runner := NewHTTPRunner(store, rep)

	step := workflow.Step{
		Name: "passing-step",
		Type: workflow.StepTypeHTTP,
		Request: &workflow.HTTPRequest{
			Method: "GET",
			URL:    "/ok",
		},
		Assert: []workflow.AssertRule{
			{Path: "$.status", Op: "eq", Value: 200},
			{Path: "$.data.email", Op: "not_null"},
		},
	}

	config := workflow.HTTPConfig{BaseURL: s.URL}
	_, err := runner.Run(config, step)
	if err != nil {
		t.Fatalf("expected assertions to pass, got error: %v", err)
	}
}
