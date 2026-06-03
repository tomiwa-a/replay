---
agent_instructions: true
---

# Replay — Contributor Guide

You are an expert Go engineer contributing to the Replay project.

## Project Structure

```
replay/
├── cmd/               # Cobra CLI commands (run, validate, install, watch, version)
├── internal/
│   ├── ai/            # AI skill definitions + installer (embedded skill content)
│   ├── config/        # Configuration file loading
│   ├── engine/        # Workflow execution engine (step execution, state, call stack)
│   ├── functions/     # Template helper functions
│   ├── parser/        # YAML workflow file parser
│   ├── reporter/      # Output formatting (JUnit, TAP, etc.)
│   ├── runner/        # Step runners (HTTP, DB, Command)
│   ├── state/         # Shared state management across steps
│   ├── template/      # Template engine (variables, expressions)
│   ├── validate/      # Workflow validation + cycle detection
│   └── workflow/      # Workflow/Step Go structs
├── plan/              # Product plan, TRD, entities, tasks
└── build/             # Built binary output
```

## Build & Test

- Build: `go build -o build/replay .`
- Test all: `go test ./...`
- Test single package: `go test ./internal/engine/... -v`
- Lint: `golangci-lint run`

## Conventions

- **Errors**: Always use `fmt.Errorf("context: %w", err)` for wrapping.
- **Testing**: Table-driven tests with `t.Run()` subtests.
- **Naming**: `snake_case` for YAML fields, `camelCase` for Go exports.
- **Imports**: Standard lib → third-party → internal, grouped with blank lines.
- **Step types**: Defined in `internal/workflow/workflow.go` as `StepType*` constants.
- **Runners**: Each step type gets a runner in `internal/runner/` implementing the `Runner` interface.

## Key Design Decisions

- **Flat step list**: Workflows are flat arrays of steps, not nested. Control flow via `if/else/loop`.
- **Call stack**: Runtime cycle detection uses `file:step` pairs. Same pair in stack = cycle.
- **Max call depth**: Default 100, configurable via `--max-call-depth`.
- **State**: Shared across steps in a workflow via `$.steps.<name>.response.body`.
- **Templates**: Go `text/template` based with custom functions.
