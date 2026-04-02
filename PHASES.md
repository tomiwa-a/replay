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
Status: **Next** 🚀

- [ ] **Workflow Runner Engine**
    - [ ] Create a central `Engine` that orchestrates steps
    - [ ] Implement sequential step execution logic
    - [ ] Integrate State Bag lifecycle (create once per run)
- [ ] **CLI Run Command**
    - [ ] Add `replay run <file>` command via Cobra
    - [ ] Logic to load, validate, and execute a workflow
    - [ ] Initial basic terminal logging (step start/end)
- [ ] **Environment Support**
    - [ ] Support loading `.env` files and mapping to ${VAR} in YAML
- [ ] **Integration Tests**
    - [ ] End-to-end test of `replay run` with a local mock server

---

## Phase 4: Shell & Database Integration
Status: **Planned** 📅

- [ ] **Shell Runner (The "Any App" Runner)**
    - [ ] Implement `type: shell` to execute CLI tools (`docker`, `sqlcmd`, etc.)
    - [ ] Capture `stdout` and `stderr` for extraction/assertion
- [ ] **PostgreSQL Native Adapter**
    - [ ] Integrate `pgx` for connection pooling
    - [ ] Implement raw SQL query execution
    - [ ] Support row-to-JSON mapping for extraction
- [ ] **Redis Native Adapter**
    - [ ] Integrate `go-redis`
    - [ ] Implement basic commands (`SET`, `GET`, `DEL`, `EXISTS`)

---

## Phase 5: Assertions & Reporting
Status: **Planned** 📅

- [ ] **Assertion Engine**
    - [ ] Implement operators: `eq`, `ne`, `gt`, `lt`, `contains`, `not_null`
    - [ ] Support JSONPath targeting in assertions
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
- [ ] **Cross-file Imports**
    - [ ] Ability to `include` steps from other workflow files
- [ ] **Release Preparation**
    - [ ] GitHub Actions for cross-platform builds (Linux, macOS, Windows)
    - [ ] Documentation for Open Source contributors
