# Research: Client Routing System

**Feature**: 002-client-routing
**Date**: 2026-02-25

## R1: Config path reconciliation

**Decision**: Use existing path `~/.config/agent-pulse/config.yaml`
**Rationale**: Feature 001 already uses this path in `internal/config/config.go`. Changing it would break existing installations. The functional spec mentioned `~/.agent-pulse/config.yaml` but the tech spec and existing code agree on `~/.config/`.
**Alternatives considered**: XDG Base Directory spec (`$XDG_CONFIG_HOME/agent-pulse/`). Rejected because the current hardcoded path is simpler and consistent with existing behavior.

## R2: Hook mechanism — Go binary vs curl/jq

**Decision**: Replace curl/jq hooks with `agent-pulse hook <event>` cobra subcommands
**Rationale**: The tech spec explicitly requires this. Benefits: typed stdin parsing, no external tool dependencies (curl, jq), consistent error handling, self-healing logic in Go rather than bash.
**Alternatives considered**: Keep curl hooks and add client routing server-side only. Rejected because jq dependency is fragile and the self-healing bash logic is hard to maintain.

## R3: Dispatcher concurrency model

**Decision**: Goroutine fan-out with `sync.WaitGroup`, per-client `http.Client` with configured timeout
**Rationale**: Simple, idiomatic Go. Each client gets a goroutine, all run in parallel, WaitGroup ensures all complete (or timeout) before handler returns. No need for worker pools at the expected scale (≤20 clients).
**Alternatives considered**: Channel-based worker pool. Rejected as over-engineered for ≤20 clients. Async fire-and-forget without WaitGroup. Rejected because we want logging of all outcomes before returning 202.

## R4: Event payload structure

**Decision**: Keep existing `Event{Type, Data}` struct. Hook subcommands populate `Data` with per-event-type fields from stdin. Dispatcher forwards the full event JSON to clients.
**Rationale**: The existing struct uses `json.RawMessage` for Data, which is flexible enough. The tech spec defines these Data fields per event type:
- `session_start`: `session_id`
- `stop`: `message`, `session_id`
- `notification`: `message`, `notification_type`, `session_id`

The functional spec wanted a flat `type, timestamp, session_id, project` payload. The tech spec enriches this with `data` containing event-specific fields. We follow the tech spec since it's the technical direction.

**Alternatives considered**: Flatten all fields at top level. Rejected because the `type` + `data` pattern is clean and extensible.

## R5: Interactive wizard library

**Decision**: Use `bufio.Scanner` for simple stdin prompting (no external library)
**Rationale**: The wizard has straightforward prompts (text input, yes/no, multi-select from 3 options). A full TUI library (bubbletea, survey) adds dependency weight for minimal benefit. The tech spec's wizard example shows simple `?` prompts.
**Alternatives considered**: `github.com/AlecAivazis/survey/v2` — feature-rich but archived. `github.com/charmbracelet/bubbletea` — excellent but adds significant dependency tree for simple prompts.

## R6: Client reachability check for `client list`

**Decision**: Removed. `client list` displays configuration only — no HTTP health checks.
**Rationale**: The list command should be a fast, offline operation that reads config and prints it. Reachability checks add latency, network dependencies, and complexity to what should be a simple config display.

## R7: Duplicate server detection

**Decision**: Try to bind the port. If "address already in use" error → server already running. No PID file needed.
**Rationale**: The existing code already handles this case in `server.go` with a user-friendly error message. For self-healing in `hook --provider claude --event session_start`, we check `/health` endpoint — if it responds, server is running.
**Alternatives considered**: PID file at `~/.config/agent-pulse/agent-pulse.pid`. Rejected because stale PID files require cleanup logic and the port-based check is simpler and more reliable.

## R8: WebSocket removal

**Decision**: Remove WebSocket hub entirely. Delete `websocket.go` and remove `gorilla/websocket` dependency.
**Rationale**: HTTP POST fan-out to registered clients replaces the WebSocket broadcast model. WebSocket required clients to maintain persistent connections; HTTP POST is simpler, works with stateless endpoints (webhooks, ESP32s), and aligns with the client routing architecture. No known consumers depend on the WebSocket endpoint.
**Alternatives considered**: Keep WebSocket alongside HTTP dispatcher. Rejected — two delivery mechanisms adds complexity with no clear benefit.

## R9: Bind address configuration

**Decision**: Add `bind_address` field to config YAML (default: `127.0.0.1`). Server uses this for `net.Listen`.
**Rationale**: Functional spec clarification confirmed configurable bind address. Default to localhost for security; users can set `0.0.0.0` for network access.
**Alternatives considered**: CLI flag `--bind`. Could add later, but YAML config is primary.

## R10: ESP32 firmware scope

**Decision**: ESP32 firmware is out of scope for this Go feature. It's a separate project that consumes events via HTTP POST.
**Rationale**: The tech spec includes ESP32 details for context (it's the primary physical client use case), but the implementation scope is the Go CLI and server.
**Alternatives considered**: Include ESP32 Arduino code in this repo. Rejected — separate hardware project with different toolchain.
