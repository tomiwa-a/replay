# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- **Version package** with `replay version` command and ldflags injection
- **Dockerfile** with multi-stage alpine build (~25MB image)
- **GoReleaser Docker** publishing to GHCR (multi-arch: amd64/arm64)
- **Makefile** with build, test, lint, docker, release targets
- **MIT License**
- **CI linting** with golangci-lint
- **CHANGELOG.md**

## [0.1.1] - 2026-01-01

### Added
- Initial release with core workflow engine
- HTTP, Shell, DB (PostgreSQL/Redis), Print step types
- Loop, Call, If step types
- Variable interpolation with raymond template engine
- JSONPath extraction and assertions
- Worker pool for multi-workflow execution
- `--concurrency`, `--fail-fast`, `--debug` flags
- `include` directive with parameter passing
- Parallel shell execution
- Scoped state (global, workflow, step)
- Call step isolation with `returns` field
- Variable validation (type, required, pattern, default)
- Step cleanup via `cleanup` field
- TTL-based variable expiration
- 45 built-in template functions (string, math, date, JSON, type)
- Config file support (`replay.yaml` with profiles and presets)
- Watch mode (`replay watch` with debounce)
- JSON Schema for workflow validation
