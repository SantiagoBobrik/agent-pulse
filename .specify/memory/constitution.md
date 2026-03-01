<!--
Sync Impact Report
- Version change: 0.0.0 → 1.0.0 (initial ratification)
- Added principles:
  1. Idiomatic Go
  2. HTTP-First Delivery
  3. Graceful Degradation
  4. Structured Observability
  5. Test Discipline
- Added sections: Technology Constraints, Development Workflow
- Templates requiring updates:
  - .specify/templates/plan-template.md — ✅ Constitution Check already present
  - .specify/templates/spec-template.md — ✅ No changes needed
  - .specify/templates/tasks-template.md — ✅ No changes needed
  - .specify/templates/agent-file-template.md — ✅ No changes needed
- Follow-up TODOs: none
-->

# agent-pulse Constitution

## Core Principles

### I. Idiomatic Go

All code MUST follow standard Go conventions for the project's minimum
version (Go 1.24+). Specific rules:

- Use `any` instead of `interface{}`.
- Use `slices.Contains` (and other stdlib generics) instead of manual
  for-range membership loops.
- Wrap errors with context: `fmt.Errorf("action: %w", err)`.
- Cobra commands MUST use `RunE` (not `Run`); propagate errors, never
  `os.Exit` in library code.
- No `fmt.Print` in library packages (`internal/`); use `slog` for
  structured output. `fmt.Print` is allowed only in `cmd/`.

**Rationale**: Consistency across contributors; leverage stdlib over
third-party when the stdlib solution is adequate.

### II. HTTP-First Delivery

Event delivery to clients MUST use HTTP POST with JSON payloads. Design
decisions:

- Clients are HTTP endpoints, not WebSocket consumers.
- Delivery is fire-and-forget: log failures, never queue or retry.
- One client's failure MUST NOT block delivery to others (concurrent
  dispatch via goroutines + `sync.WaitGroup`).
- Authentication limited to bearer tokens with `${ENV_VAR}` resolution
  at dispatch time.

**Rationale**: HTTP endpoints are the lowest-common-denominator
integration surface (webhooks, lambdas, ESPs). Simplicity over
guaranteed delivery — consumers that need reliability can add their own
ack layer.

### III. Graceful Degradation

The system MUST remain operational under partial failure:

- Missing config file → return sensible defaults (port 8080,
  bind 127.0.0.1), never error.
- Invalid config file → fail loudly with descriptive error and path.
- Unreachable client → log and skip, continue delivering to others.
- Server not running when hook fires → auto-start via health check,
  poll up to 3 times over 3 seconds.
- All domain models MUST expose a `Validate()` method that returns a
  descriptive error; validation happens after parse, before persist.

**Rationale**: agent-pulse runs as background infrastructure. Silent
crashes or hard failures on edge cases erode trust.

### IV. Structured Observability

All runtime behavior MUST be observable through structured logging:

- Use stdlib `slog` exclusively — no third-party loggers.
- Log key-value pairs: `slog.Info("event received", "type", ev.Type)`.
- Errors MUST include enough context to diagnose without a debugger:
  client name, URL, status code, event type.
- No silent swallowing of errors; if an error is intentionally ignored,
  add a comment explaining why.

**Rationale**: In a fire-and-forget system, logs are the primary
debugging tool. Structured logs enable machine parsing for future
dashboards.

### V. Test Discipline

Tests MUST follow these conventions:

- Standard `testing` package — no third-party test frameworks.
- Table-driven tests for validation and parsing logic.
- Temporary directories via `t.TempDir()`; isolate HOME with
  `t.Setenv("HOME", dir)` to prevent side effects.
- Error expectation uses `wantErr bool` pattern:
  `if (err != nil) != tt.wantErr`.
- Test names follow `TestFunctionName` with descriptive subtests.
- NEVER delete existing tests unless explicitly asked.

**Rationale**: Deterministic, isolated tests that run fast and never
flake. Table-driven style keeps validation tests readable as cases grow.

## Technology Constraints

The following technology choices are locked for the current project
phase and MUST NOT be changed without a constitution amendment:

| Layer        | Choice                                    |
|--------------|-------------------------------------------|
| Language     | Go 1.24+                                  |
| CLI          | cobra                                     |
| HTTP router  | chi/v5                                    |
| Config       | gopkg.in/yaml.v3 (`~/.config/agent-pulse/config.yaml`) |
| Logging      | stdlib `slog`                             |
| Auth         | Bearer token with env-var substitution    |
| Delivery     | HTTP POST, fire-and-forget                |

Adding a new dependency MUST be justified by a gap in the stdlib or
existing deps. Prefer stdlib solutions.

## Development Workflow

All changes MUST follow this workflow:

1. **Pre-commit**: Run `go vet ./...` and `go test ./...` locally
   before every commit. Never rely on CI to catch issues.
2. **Commit messages**: Clean, user-attributed. No co-author trailers.
3. **Pull requests**: Never create without explicit user confirmation.
   Show title, summary, and target branch first.
4. **Feature work**: Follow speckit flow — spec → clarify → plan →
   tasks → implement. Each user story MUST be independently testable.
5. **Config changes**: Bump defaults only with backward-compatible
   migration. Document in CLAUDE.md.

## Governance

This constitution is the authoritative source for project conventions.
All code reviews and plan validations MUST verify compliance with these
principles.

**Amendment procedure**:

1. Propose change with rationale in a PR or conversation.
2. Update this file with the change.
3. Increment version per semver:
   - MAJOR: principle removal or incompatible redefinition.
   - MINOR: new principle or materially expanded guidance.
   - PATCH: clarification, wording, typo fix.
4. Update `LAST_AMENDED_DATE`.
5. Run consistency propagation across templates.

**Compliance review**: Every `/speckit.plan` invocation MUST pass the
Constitution Check gate before proceeding to Phase 0 research.

Use `CLAUDE.md` for runtime development guidance that complements this
constitution.

**Version**: 1.0.0 | **Ratified**: 2026-02-26 | **Last Amended**: 2026-02-26
