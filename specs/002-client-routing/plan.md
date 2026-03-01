# Implementation Plan: Client Routing System

**Branch**: `002-client-routing` | **Date**: 2026-02-25 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/002-client-routing/spec.md`

## Summary

Extend the existing event bridge (feature 001) to replace WebSocket broadcasting with HTTP POST fan-out to registered clients. This feature adds a client registry (YAML config), a dispatcher that fans out events as HTTP POST requests with per-client filtering/timeout/auth, CLI commands for client management (`client add/list/remove`), and Go-native hook subcommands that replace the fragile curl/jq shell hooks. The WebSocket hub is removed entirely — fan-out to HTTP clients is the delivery mechanism.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: cobra (CLI), chi/v5 (HTTP router), gopkg.in/yaml.v3 (config), slog (stdlib logging)
**Storage**: YAML file at `~/.config/agent-pulse/config.yaml` (existing path)
**Testing**: Table-driven tests + `net/http/httptest`
**Target Platform**: macOS / Linux (developer workstation)
**Project Type**: CLI tool + background HTTP server
**Performance Goals**: Event delivery to all clients within 1s; fan-out must not let one client block others
**Constraints**: Fire-and-forget delivery (no retries); per-client timeout default 2s
**Scale/Scope**: Up to 20 simultaneous clients

## Constitution Check

*No constitution file found — gate passes by default.*

## Project Structure

### Documentation (this feature)

```text
specs/002-client-routing/
├── plan.md              # This file
├── research.md          # Phase 0 output
├── data-model.md        # Phase 1 output
├── quickstart.md        # Phase 1 output
├── contracts/           # Phase 1 output
│   └── http-api.md      # POST /event contract
└── tasks.md             # Phase 2 output (/speckit.tasks)
```

### Source Code (repository root)

```text
agent-pulse/
├── main.go                          # Entry point (exists)
├── go.mod                           # Module (exists)
├── cmd/
│   ├── root.go                      # Root cobra command (exists)
│   ├── serve.go                     # serve command (exists — modify to init dispatcher)
│   ├── setup.go                     # setup command (exists — modify hook format)
│   ├── hook.go                      # NEW — hook parent + session_start/stop/notification subcommands
│   └── client.go                    # NEW — client parent + add/list/remove subcommands
├── internal/
│   ├── config/
│   │   └── config.go                # Config loader (exists — extend with clients + bind_address)
│   ├── server/
│   │   ├── server.go                # HTTP server (exists — add dispatcher integration)
│   │   ├── handler.go               # Event handler (exists — wire dispatcher instead of hub)
│   │   └── dispatcher.go            # NEW — fan-out to HTTP clients (replaces websocket.go)
│   ├── hooks/
│   │   └── setup.go                 # Hook injection (exists — change to Go binary hooks)
│   └── client/
│       ├── client.go                # NEW — Client model + validation
│       └── wizard.go                # NEW — Interactive wizard for client add
```

**Structure Decision**: Extend the existing `cmd/` + `internal/` layout from feature 001. New packages: `internal/client/` for the client model and wizard. New files in existing packages: `cmd/hook.go`, `cmd/client.go`, `internal/server/dispatcher.go`.

## Key Design Decisions

### 1. Hook commands replace curl/jq

The existing hooks in `.claude/settings.json` use `curl` and `jq` shell commands. This feature replaces them with `agent-pulse hook <event>` subcommands that read stdin JSON natively in Go. This eliminates the dependency on curl/jq and provides typed parsing.

### 2. Dispatcher replaces WebSocket hub

The existing WebSocket broadcast hub (`websocket.go`) is removed. The new HTTP POST dispatcher is the sole delivery mechanism. `gorilla/websocket` dependency is also removed from `go.mod`.

### 3. Config path stays at `~/.config/agent-pulse/config.yaml`

The existing code uses this path. The functional spec mentioned `~/.agent-pulse/config.yaml` but we follow the established convention.

### 4. Event payload: type + data (existing)

The existing `Event` struct has `type` and `data` (json.RawMessage). The hook subcommands populate `data` with event-specific fields from Claude Code's stdin JSON (`session_id`, `message`, `notification_type`). The dispatcher forwards the full event to clients.

### 5. Self-healing moves to Go

The session_start hook currently uses bash to check health and start the server. The new `agent-pulse hook --provider claude --event session_start` command does this in Go: check `/health`, if unreachable start `agent-pulse serve` as a background process, wait for ready, then dispatch.

## Complexity Tracking

No constitution violations — table not applicable.
