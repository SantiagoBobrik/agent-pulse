# Research: Event Bridge

**Date**: 2026-02-25
**Branch**: `001-event-bridge`

## 1. CLI Framework (cobra)

**Decision**: Use cobra with one file per command in `cmd/` package. `main.go` at root calls `cmd.Execute()`.

**Rationale**: cobra is the Go standard for CLI tools. The one-file-per-command pattern keeps the cmd package thin and testable. Business logic lives in `internal/` packages, not in cobra `RunE` functions.

**Alternatives considered**:
- `urfave/cli` — viable but less ecosystem adoption, fewer examples
- Plain `flag` — too low-level for multi-command CLI
- Custom `os.Args` parsing — unnecessary complexity

**Key decisions**:
- Use `RunE` (not `Run`) for error propagation
- Use `PersistentFlags()` on root for shared flags (e.g., `--port`)
- No business logic in command files — delegate to `internal/` packages

## 2. HTTP Server (chi)

**Decision**: Use chi with a flat route setup. Two endpoints only: `POST /event` and `GET /health`.

**Rationale**: chi is lightweight, `net/http` compatible, and has built-in middleware including `middleware.Heartbeat("/health")` which handles the health endpoint automatically.

**Alternatives considered**:
- `net/http` ServeMux — viable for 2 routes but chi adds logging/recovery middleware for free
- `gin` — heavier, unnecessary for this scope
- `echo` — similar to gin, overkill

**Key decisions**:
- Use `middleware.Heartbeat("/health")` for the health check (no custom handler needed)
- Use `middleware.Recoverer` to prevent panics from crashing the server
- Handler functions receive dependencies via closures (not globals)
- Return `http.Handler` interface from router constructor

## 3. WebSocket Client Pool (gorilla/websocket)

**Decision**: Use the canonical gorilla Hub pattern — a single goroutine owns the client map, communicating via channels. Each client gets `readPump` and `writePump` goroutines.

**Rationale**: This pattern is battle-tested, avoids mutexes entirely (single goroutine owns state), and handles concurrent access safely.

**Alternatives considered**:
- `sync.RWMutex` on client map — more error-prone, race conditions harder to catch
- `nhooyr.io/websocket` — newer but less ecosystem support, fewer examples for hub pattern
- Single goroutine per client — doesn't scale for broadcast pattern

**Key decisions**:
- Hub communicates via `register`, `unregister`, and `broadcast` channels
- Client `send` channel is buffered (256) — if full, client is assumed dead and removed
- Implement ping/pong with deadlines to detect stale connections
- `CheckOrigin` returns `true` (local dev tool, no CORS restrictions needed)
- WebSocket close: send close frame, then close TCP connection

## 4. Claude Code Hooks

**Decision**: Configure hooks in `.claude/settings.json` per-project. Use `curl -d @-` to pipe stdin JSON directly to the HTTP server.

**Rationale**: Hooks receive JSON on stdin with event data. The simplest approach is piping directly to the local server. The SessionStart hook additionally checks server health and auto-starts if needed.

**Key findings**:
- All hooks receive: `session_id`, `transcript_path`, `cwd`, `hook_event_name`
- **SessionStart** extra: `source` (startup/resume/clear/compact), `model`
- **Stop** extra: `stop_hook_active` (bool), `last_assistant_message` (string — but we're not using this per clarification)
- **Notification** extra: `message`, `title`, `notification_type` (permission_prompt/idle_prompt/auth_success/elicitation_dialog)
- Exit code 0 = success, exit code 2 = blocking error
- Hooks snapshot at Claude Code startup — changes require restart

**Hook command patterns**:
- SessionStart: health check → auto-start → POST event
- Stop: pipe stdin to POST (but we only send `{type: "stop"}`)
- Notification: pipe stdin with jq transform to POST

## 5. JSON Settings Merge

**Decision**: Use `map[string]interface{}` for generic JSON merge. Read → unmarshal → merge hooks key → marshal with indent → write.

**Rationale**: Using a generic map preserves all unknown keys. A typed struct would drop fields not in the struct definition.

**Key decisions**:
- If file doesn't exist, start with empty map and create `.claude/` directory
- If file has invalid JSON, return error (don't silently overwrite)
- Use `json.MarshalIndent` with 2-space indent
- Append trailing newline
- Consider atomic write (write temp file → rename) for crash safety

## 6. Graceful Shutdown

**Decision**: Use `signal.NotifyContext` for SIGINT/SIGTERM. Close WebSocket connections first, then `http.Server.Shutdown()` with timeout.

**Rationale**: `http.Server.Shutdown()` doesn't close hijacked (WebSocket) connections. They must be closed explicitly before HTTP shutdown.

**Key decisions**:
- 5-second shutdown timeout
- Order: close WebSockets → shutdown HTTP server
- Set `ReadHeaderTimeout: 5s` to prevent slowloris
- Don't call `os.Exit()` — let `main()` return naturally for deferred cleanup
- Double Ctrl+C force-quits (default behavior after first signal consumed)
