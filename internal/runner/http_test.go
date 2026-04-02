package runner

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
	runner := NewHTTPRunner(store)

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

	_, err := runner.Run(s.URL, step)
	if err != nil {
		t.Fatalf("Run failed: %v", err)
	}

	token, ok := store.Get("my_token")
	if !ok || token != "secret-token" {
		t.Errorf("expected secret-token extracted, got %v", token)
	}
}
