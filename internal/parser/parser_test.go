package parser

import (
	"strings"
	"testing"
)

func TestLoadFromBytes_ValidWorkflow(t *testing.T) {
	data := []byte(`
name: test-workflow
steps:
  - name: login
    type: http
    request:
      method: POST
      url: /auth/login
`)

	wfs, err := LoadFromBytes(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(wfs) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(wfs))
	}

	wf := wfs[0]
	if wf.Name != "test-workflow" {
		t.Fatalf("expected name test-workflow, got %q", wf.Name)
	}

	if wf.Name != "test-workflow" {
		t.Fatalf("expected name test-workflow, got %q", wf.Name)
	}

	if wf.Version != "v0.1" {
		t.Fatalf("expected default version v0.1, got %q", wf.Version)
	}

	if len(wf.Steps) != 1 {
		t.Fatalf("expected 1 step, got %d", len(wf.Steps))
	}
}

func TestLoadFromBytes_MultiWorkflow(t *testing.T) {
	data := []byte(`
name: flow-a
steps:
  - name: s1
    type: print
    message: a
---
name: flow-b
steps:
  - name: s2
    type: print
    message: b
`)

	wfs, err := LoadFromBytes(data)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(wfs) != 2 {
		t.Fatalf("expected 2 workflows, got %d", len(wfs))
	}

	if wfs[0].Name != "flow-a" {
		t.Errorf("expected flow-a, got %s", wfs[0].Name)
	}
	if wfs[1].Name != "flow-b" {
		t.Errorf("expected flow-b, got %s", wfs[1].Name)
	}
}

func TestLoadFromBytes_UnknownFieldFails(t *testing.T) {
	data := []byte(`
name: invalid-workflow
unknown_field: value
steps:
  - name: s1
    type: http
    request:
      method: GET
      url: /health
`)

	_, err := LoadFromBytes(data)
	if err == nil {
		t.Fatal("expected error for unknown field, got nil")
	}

	if !strings.Contains(err.Error(), "unknown_field") {
		t.Fatalf("expected error to mention unknown field, got %v", err)
	}
}

func TestLoadFromBytes_MalformedYAMLFails(t *testing.T) {
	data := []byte(`
name: bad
steps:
  - name: s1
    type http
`)

	_, err := LoadFromBytes(data)
	if err == nil {
		t.Fatal("expected malformed yaml error, got nil")
	}
}
