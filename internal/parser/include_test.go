package parser

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveIncludes_NoIncludes(t *testing.T) {
	tmpDir := t.TempDir()
	mainPath := filepath.Join(tmpDir, "main.yaml")
	mainContent := `name: test
steps:
  - name: step1
    message: "hello"
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatal(err)
	}

	wf, err := LoadFromFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}

	if len(wf) != 1 {
		t.Fatalf("expected 1 workflow, got %d", len(wf))
	}

	if err := ResolveIncludes(&wf[0]); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(wf[0].Steps) != 1 {
		t.Errorf("expected 1 step, got %d", len(wf[0].Steps))
	}
}

func TestResolveIncludes_SingleFile(t *testing.T) {
	tmpDir := t.TempDir()

	incPath := filepath.Join(tmpDir, "auth.yaml")
	incContent := `name: auth-steps
steps:
  - name: login
    message: "logging in as {{ username }}"
  - name: verify
    message: "verifying token"
`
	if err := os.WriteFile(incPath, []byte(incContent), 0644); err != nil {
		t.Fatal(err)
	}

	mainPath := filepath.Join(tmpDir, "main.yaml")
	mainContent := `name: main
include:
  - file: auth.yaml
    with:
      username: alice
steps:
  - name: main-step
    message: "running main"
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatal(err)
	}

	wf, err := LoadFromFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := ResolveIncludes(&wf[0]); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(wf[0].Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(wf[0].Steps))
	}

	if wf[0].Steps[0].Message != "logging in as alice" {
		t.Errorf("expected param substitution in step 0, got %q", wf[0].Steps[0].Message)
	}
	if wf[0].Steps[1].Message != "verifying token" {
		t.Errorf("expected step 1 unchanged, got %q", wf[0].Steps[1].Message)
	}
	if wf[0].Steps[2].Message != "running main" {
		t.Errorf("expected main step at end, got %q", wf[0].Steps[2].Message)
	}
}

func TestResolveIncludes_MissingFile(t *testing.T) {
	tmpDir := t.TempDir()
	mainPath := filepath.Join(tmpDir, "main.yaml")
	mainContent := `name: main
include:
  - file: nonexistent.yaml
steps:
  - name: step1
    message: "hello"
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatal(err)
	}

	wf, err := LoadFromFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}

	err = ResolveIncludes(&wf[0])
	if err == nil {
		t.Error("expected error for missing include file, got nil")
	}
}

func TestResolveIncludes_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	authPath := filepath.Join(tmpDir, "auth.yaml")
	authContent := `name: auth
steps:
  - name: login
    message: "login as {{ user }}"
`
	if err := os.WriteFile(authPath, []byte(authContent), 0644); err != nil {
		t.Fatal(err)
	}

	setupPath := filepath.Join(tmpDir, "setup.yaml")
	setupContent := `name: setup
steps:
  - name: init
    message: "init {{ env }}"
`
	if err := os.WriteFile(setupPath, []byte(setupContent), 0644); err != nil {
		t.Fatal(err)
	}

	mainPath := filepath.Join(tmpDir, "main.yaml")
	mainContent := `name: main
include:
  - file: auth.yaml
    with:
      user: bob
  - file: setup.yaml
    with:
      env: prod
steps:
  - name: final
    message: "done"
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatal(err)
	}

	wf, err := LoadFromFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := ResolveIncludes(&wf[0]); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(wf[0].Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(wf[0].Steps))
	}

	if wf[0].Steps[0].Message != "login as bob" {
		t.Errorf("expected 'login as bob', got %q", wf[0].Steps[0].Message)
	}
	if wf[0].Steps[1].Message != "init prod" {
		t.Errorf("expected 'init prod', got %q", wf[0].Steps[1].Message)
	}
	if wf[0].Steps[2].Message != "done" {
		t.Errorf("expected 'done', got %q", wf[0].Steps[2].Message)
	}
}

func TestResolveIncludes_Nested(t *testing.T) {
	tmpDir := t.TempDir()

	innerPath := filepath.Join(tmpDir, "inner.yaml")
	innerContent := `name: inner
steps:
  - name: inner-step
    message: "from inner: {{ x }}"
`
	if err := os.WriteFile(innerPath, []byte(innerContent), 0644); err != nil {
		t.Fatal(err)
	}

	outerPath := filepath.Join(tmpDir, "outer.yaml")
	outerContent := `name: outer
include:
  - file: inner.yaml
steps:
  - name: outer-step
    message: "from outer"
`
	if err := os.WriteFile(outerPath, []byte(outerContent), 0644); err != nil {
		t.Fatal(err)
	}

	mainPath := filepath.Join(tmpDir, "main.yaml")
	mainContent := `name: main
include:
  - file: outer.yaml
    with:
      x: nested-value
steps:
  - name: main-step
    message: "main"
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatal(err)
	}

	wf, err := LoadFromFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}

	if err := ResolveIncludes(&wf[0]); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(wf[0].Steps) != 3 {
		t.Fatalf("expected 3 steps (inner + outer + main), got %d", len(wf[0].Steps))
	}
}

func TestResolveIncludes_MissingFileField(t *testing.T) {
	tmpDir := t.TempDir()
	mainPath := filepath.Join(tmpDir, "main.yaml")
	mainContent := `name: main
include:
  - with:
      foo: bar
steps:
  - name: step1
    message: "hello"
`
	if err := os.WriteFile(mainPath, []byte(mainContent), 0644); err != nil {
		t.Fatal(err)
	}

	wf, err := LoadFromFile(mainPath)
	if err != nil {
		t.Fatal(err)
	}

	err = ResolveIncludes(&wf[0])
	if err == nil {
		t.Error("expected error for missing 'file' field, got nil")
	}
}
