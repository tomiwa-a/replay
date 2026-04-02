# Implementation Roadmap: Project Replay

This document tracks the detailed engineering tasks for Replay.

## Phase 1: DSL and Parser

Status: **Completed** ✅

- [x] Define Phase 1 YAML Schema (name, steps, http/db blocks)
- [x] Create Go struct model for Workflow and Steps
- [x] Implement strict YAML parser with unknown-field rejection
- [x] Build multi-error validator (schema validation, field requirements)
- [x] Create basic CLI with `validate` command
- [x] Write parser and validator unit tests
- [x] Create sample valid and invalid workflow files

---

## Phase 2: State Management & HTTP Runner

Status: **Completed** ✅

- [x] **State Bag Implementation**
  - [x] Create concurrency-safe variable store
  - [x] Implement `Get` / `Set` for dynamic values
- [x] **Templating Engine**
  - [x] Build regex-based interpolator for `{{ var }}` strings
  - [x] Support recursive placeholder resolution
- [x] **HTTP Runner Core**
  - [x] Implement `http` step executor
  - [x] Support GET, POST, PUT, DELETE with headers and body
  - [x] Add basic JSON response parsing
- [x] **Data Extraction**
  - [x] Integrate a JSONPath library (`github.com/ohler55/ojg`)
  - [x] Implement `extract` logic to save response fields to State Bag
- [x] **Phase 2 Tests**
  - [x] Mock HTTP server tests
  - [x] Variable interpolation unit tests

---

## Phase 3: Execution Engine & CLI

Status: **Completed** ✅

- [x] **Workflow Runner Engine**
  - [x] Create a central `Engine` that orchestrates steps
  - [x] Implement sequential step execution logic
  - [x] Integrate State Bag lifecycle (create once per run)
- [x] **CLI Run Command**
  - [x] Add `replay run <file>` command via Cobra
  - [x] Logic to load, validate, and execute a workflow
  - [x] Initial basic terminal logging (step start/end)
- [x] **Environment Support**
  - [x] Support loading `.env` files and mapping to ${VAR} in YAML
- [x] **Integration Tests**
  - [x] End-to-end test of `replay run` with a local mock server

---

## Phase 4: Shell & Database Integration

Status: **Completed** ✅

- [x] **Shell Runner (The "Any App" Runner)**
  - [x] Implement `type: shell` to execute CLI tools (`docker`, `sqlcmd`, etc.)
  - [x] Capture `stdout` and `stderr` for extraction/assertion
  - [x] Add execution timeout support (e.g., `timeout: 30s`)
  - [x] Support multi-command sequential execution (list of commands)
- [x] **PostgreSQL Native Adapter**
  - [x] Integrate `pgx` for connection pooling
  - [x] Implement raw SQL query execution
  - [x] Support row-to-JSON mapping for extraction
- [x] **Redis Native Adapter**
  - [x] Integrate `go-redis`
  - [x] Implement basic commands (`SET`, `GET`, `DEL`, `EXISTS`)
- [x] **DB DSL Enhancements**
  - [x] Implement flat DB shortcuts (`query`, `command`, `engine` at step level)
  - [x] Default `engine` to `postgres` if not specified
- [ ] **State persistence and isolation**
  - [ ] Ensure DB results can be referenced accurately in subsequent steps

---

## Phase 5: Assertions & Reporting

Status: **In-Progress** 🏗️

- [ ] **Assertion Engine**
  - [ ] Implement operators: `eq`, `ne`, `gt`, `lt`, `contains`, `not_null`
  - [ ] Support JSONPath targeting in assertions
  - [ ] Add `ignore_error: true` support at step level
- [ ] **Clean Terminal Reporter**
  - [ ] Implement colored output (Pass/Fail)
  - [ ] Show step duration and meaningful error snippets on failure

---

## Phase 6: Parallelism & Advanced Features

Status: **Planned** 📅

- [ ] **Concurrency Engine**
  - [ ] Implement worker pool for multi-workflow execution
  - [ ] Add `--concurrency N` flag
  - [ ] Add `--fail-fast` toggle
- [ ] **Parallel Command Execution**
  - [ ] Implement `parallel: true` for shell steps with multiple commands
- [ ] **Cross-file Imports**
  - [ ] Ability to `include` steps from other workflow files
- [ ] **Release Preparation**
  - [ ] GitHub Actions for cross-platform builds (Linux, macOS, Windows)
  - [ ] Documentation for Open Source contributors
