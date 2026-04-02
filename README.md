# Replay

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://go.dev/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Status](https://img.shields.io/badge/status-active%20development-blue)](#roadmap)
[![CLI](https://img.shields.io/badge/interface-CLI-4B32C3)](#usage)
[![Open Source](https://img.shields.io/badge/open%20source-yes-brightgreen)](#contributing)

Replay is a standalone, CLI-based execution engine in Go for declarative end-to-end workflow testing.

It combines three things in one workflow file:

- HTTP calls (API verification)
- Stateful variable passing across steps (tokens, IDs, computed values)
- Native database operations (PostgreSQL and Redis)

Replay is built for teams that want fast, deterministic, and scriptable E2E validation without heavy test framework overhead.

Replay can also run multiple workflows at the same time, which is useful for CI parallelization, tenant-based testing, or high-volume regression checks.

## Why Replay

Replay is designed to make production-like test flows easy to express and reliable to run:

- Declarative YAML DSL for readable workflows
- Strong runtime performance from Go
- Native DB execution support for setup, mutation, and validation
- Step-level data extraction and assertions
- Open-source and automation-friendly by default

## Project Vision

Replay executes a YAML workflow from top to bottom. Each step can:

- Perform an action (`http` or `db`)
- Extract values into a shared state bag
- Assert expected outcomes

This lets you model complete user journeys, such as:

1. Login and capture JWT
2. Create an entity over HTTP
3. Update state directly in PostgreSQL
4. Validate final API response and database state

## Core Stack

- Language: Go
- DSL: YAML
- PostgreSQL driver: `pgx`
- Redis driver: `go-redis/redis`
- Assertions: JSONPath-driven expression checks

## Architecture

Replay is organized around these components:

1. Parser

- Loads YAML workflow files into Go structs
- Validates required fields and step shape

2. State Bag

- Concurrency-safe key-value store for runtime variables
- Shared across all workflow steps

3. Templating Layer

- Replaces placeholders such as `{{ token }}` before execution
- Applies to HTTP URL, headers, body, SQL, and command args

4. Runners

- HTTP Runner: executes GET/POST/etc., captures responses
- DB Runner: executes SQL or Redis commands using pooled connections

5. Assertion Engine

- Evaluates conditions on output payloads and metadata
- Supports JSONPath-based targeting and value comparisons

6. Reporter

- Human-friendly CLI output
- Colored pass/fail status per step and workflow summary

## Workflow DSL (Draft)

```yaml
name: user-onboarding-flow
config:
  http:
    base_url: https://api.example.com
  postgres:
    dsn: ${POSTGRES_DSN}
  redis:
    addr: ${REDIS_ADDR}

steps:
  - name: login
    type: http
    request:
      method: POST
      url: /auth/login
      headers:
        Content-Type: application/json
      body:
        email: qa@example.com
        password: pass123
    extract:
      token: $.data.token
    assert:
      - path: $.status
        op: eq
        value: 200
      - path: $.data.token
        op: not_null

  - name: verify-user-in-db
    type: db
    db:
      engine: postgres
      query: |
        SELECT id, verified_at
        FROM users
        WHERE email = 'qa@example.com';
    extract:
      user_id: $.rows[0].id
    assert:
      - path: $.rows[0].verified_at
        op: not_null

  - name: cache-session
    type: db
    db:
      engine: redis
      command: ["SET", "session:{{ user_id }}", "{{ token }}"]
    assert:
      - path: $.ok
        op: eq
        value: true
```

## Usage

Planned CLI structure:

```bash
replay run workflow.yaml
```

Future command examples:

```bash
replay validate workflow.yaml
replay run workflow.yaml --env .env --verbose
replay run workflows/*.yaml --concurrency 4
replay version
```

## Parallel Workflow Execution

Yes, this model works well in Go.

Suggested approach:

- Run each workflow in its own isolated execution context (its own state bag)
- Use a worker pool with configurable concurrency (for example, `--concurrency 4`)
- Share DB/HTTP connection pools, but isolate runtime variables per workflow
- Emit grouped logs per workflow, then print a final aggregate summary

Recommended CLI shape:

```bash
replay run workflow-a.yaml workflow-b.yaml workflow-c.yaml --concurrency 3
replay run workflows/*.yaml --concurrency 6 --fail-fast=false
```

Important safeguards:

- Do not share one state bag across workflows
- Cap max concurrency to avoid DB/API overload
- Keep deterministic output ordering in the final summary
- Support both fail-fast and continue-on-error modes

## Implementation Roadmap

### Phase 1: DSL and Parser

- Define YAML schema (`name`, `type`, `request`, `extract`, `assert`, etc.)
- Build Go struct model
- Implement parser and schema validation

### Phase 2: State and HTTP

- Implement concurrency-safe state bag
- Add placeholder templating (`{{ var }}`)
- Implement HTTP runner and extraction

### Phase 3: Database Integration

- Add connection config for PostgreSQL and Redis
- Implement PostgreSQL adapter for raw SQL execution
- Implement Redis adapter for command execution
- Support extraction from DB results into state bag

### Phase 4: Assertions and Reporting

- Implement assertion operators (`eq`, `ne`, `gt`, `contains`, `not_null`, etc.)
- Add JSONPath targeting in assertions and extraction
- Build colored terminal report and summary

### Phase 5: CLI and Packaging

- Build CLI with Cobra
- Add `replay run workflow.yaml`
- Add multi-workflow execution with `--concurrency N`
- Prepare release flow for open-source distribution

## Open Source Plan

Replay is intended to be open source from day one.

Planned project standards:

- MIT License
- Contributor guidelines
- Issue and pull request templates
- Semantic versioning
- Changelog-driven releases

## Contributing

Contributions are welcome.

Areas that will need help early:

- DSL evolution and validation rules
- Assertion engine design
- Driver adapters and integration tests
- Developer ergonomics and docs

## Status

Replay is currently in the completion of Phase 1. You can track the granular implementation progress in [PHASES.md](PHASES.md).

If this direction matches your testing needs, open an issue with your use case and expected workflow patterns.
