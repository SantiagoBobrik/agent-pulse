# Tasks: Event Bridge

**Input**: Design documents from `/specs/001-event-bridge/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Phase 1: Setup

**Purpose**: Initialize Go project and CLI skeleton

- [x] T001 Initialize Go module — run `go mod init github.com/santiagodiaz/claude-pulse`, add dependencies: `github.com/spf13/cobra`, `github.com/go-chi/chi/v5`, `github.com/gorilla/websocket`, `gopkg.in/yaml.v3` in go.mod
- [x] T002 [P] Create entry point in main.go — import `cmd` package, call `cmd.Execute()`, exit with code 1 on error
- [x] T003 [P] Create root cobra command in cmd/root.go — define `rootCmd` with Use: "claude-pulse", Short/Long descriptions per CLI contract, `Execute()` function that returns error. No `Run` func (shows help by default)

---

## Phase 2: Foundational

**Purpose**: Shared config infrastructure used by both `setup` and `serve` commands

- [x] T004 Implement config loading in internal/config/config.go — define `Config` struct with `Port int` field (yaml tag `port`), `Load()` function reads `~/.config/claude-pulse/config.yaml`, returns defaults (port 8080) if file missing, validate port range 1024-65535, use `gopkg.in/yaml.v3` for parsing

**Checkpoint**: Foundation ready — user story implementation can begin

---

## Phase 3: User Story 1 - One-Time Project Setup (Priority: P1) 🎯 MVP

**Goal**: Developer runs `claude-pulse setup` in a project directory and Claude Code hooks are configured automatically.

**Independent Test**: Run setup in a temp directory, verify `.claude/settings.json` contains correct hook entries. Run again and verify idempotency. Run with existing unrelated settings and verify preservation.

### Implementation for User Story 1

- [x] T005 [US1] Implement hook generation and JSON merge in internal/hooks/setup.go — `Setup(dir string, port int) error` function that: (1) generates hook command strings with the configured port per hooks contract (SessionStart with health check + auto-start, Stop with static JSON, Notification with jq pipe), (2) reads `.claude/settings.json` using `map[string]interface{}` for generic merge, (3) creates `.claude/` directory if missing, (4) creates settings file if missing (start with empty map), (5) returns error on invalid JSON (don't overwrite), (6) merges/replaces `hooks.SessionStart`, `hooks.Stop`, `hooks.Notification` keys preserving all other settings, (7) writes back with `json.MarshalIndent` 2-space indent + trailing newline
- [x] T006 [US1] Implement setup cobra command in cmd/setup.go — define `setupCmd` with Use: "setup", Short: "Configure Claude Code hooks for the current project", `RunE` that: (1) loads config via `config.Load()` to get port, (2) gets current working directory, (3) calls `hooks.Setup(cwd, config.Port)`, (4) prints success message: "claude-pulse hooks configured in .claude/settings.json", (5) on error prints "error: " + message and returns error. Add to rootCmd via `init()`

**Checkpoint**: `claude-pulse setup` works end-to-end. MVP is functional for configuring projects.

---

## Phase 4: User Story 3 - Real-Time Event Forwarding (Priority: P1) + User Story 2 - Automatic Server Lifecycle (Priority: P1)

**Goal**: Server receives events via HTTP and broadcasts to WebSocket-connected devices. Server starts/stops cleanly with signal handling and port-in-use detection.

**Independent Test**: Start server with `claude-pulse serve`, connect a WebSocket client (e.g., `websocat ws://localhost:8080/ws`), send `curl -X POST localhost:8080/event -d '{"type":"session_start"}'`, verify client receives the event. Send SIGINT, verify clean shutdown.

### Implementation

- [x] T007 [P] [US3] Define Event types and POST /event handler in internal/server/handler.go — define `Event` struct with `Type string` and `Extra json.RawMessage` fields (json tags: `type`, `extra`), define valid event types as constants (`session_start`, `stop`, `notification`), implement `handleEvent(hub *Hub) http.HandlerFunc` that: (1) decodes JSON body, (2) validates `Type` is one of the three allowed values, (3) marshals event back to JSON bytes, (4) sends to `hub.broadcast` channel, (5) returns 202 Accepted on success, 400 Bad Request on invalid JSON or unknown type
- [x] T008 [P] [US3] Implement WebSocket Hub and Client in internal/server/websocket.go — define `Hub` struct with `clients map[*Client]bool`, `register/unregister/broadcast` channels, `run()` method as select loop (register adds client, unregister removes + closes send channel, broadcast iterates clients with non-blocking send — full buffer means dead client, remove it). Define `Client` struct with `hub *Hub`, `conn *websocket.Conn`, `send chan []byte` (buffered 256). Implement `readPump()`: reads messages in loop (only to detect disconnect), defers `hub.unregister` and `conn.Close()`. Implement `writePump()`: reads from `send` channel, writes text message to conn. Implement `serveWs(hub *Hub) http.HandlerFunc`: upgrades HTTP to WebSocket with `CheckOrigin: func(r *http.Request) bool { return true }`, creates Client, registers with hub, starts readPump and writePump as goroutines. Export `NewHub()` constructor and `Hub.Run()`, `Hub.Shutdown()` methods
- [x] T009 [US3] [US2] Implement HTTP server in internal/server/server.go — define `Server` struct wrapping `*http.Server` and `*Hub`. `NewServer(hub *Hub, port int) *Server`: creates chi router, adds `middleware.Heartbeat("/health")`, `middleware.Recoverer`, adds `POST /event` route with `handleEvent(hub)`, adds `GET /ws` route with `serveWs(hub)`, sets `Addr: fmt.Sprintf(":%d", port)`, sets `ReadHeaderTimeout: 5*time.Second`. `Start() error`: calls `ListenAndServe`, returns error (detect port-in-use: if error contains "address already in use", return formatted error: `fmt.Errorf("port %d is already in use. Change the port in ~/.config/claude-pulse/config.yaml", port)`). `Shutdown(ctx context.Context) error`: calls `hub.Shutdown()` to close all WebSocket connections, then calls `http.Server.Shutdown(ctx)`
- [x] T010 [US3] [US2] Implement serve cobra command in cmd/serve.go — define `serveCmd` with Use: "serve", Short: "Start the event bridge server", add `--port/-p` int flag (default 0, meaning "use config"). `RunE`: (1) load config, (2) override port if flag set, (3) create Hub with `NewHub()`, (4) `go hub.Run()`, (5) create Server with `NewServer(hub, port)`, (6) start server in goroutine, (7) log `[claude-pulse] server started on :%d`, (8) set up `signal.NotifyContext` for SIGINT/SIGTERM, (9) wait for context done, (10) log `[claude-pulse] server shutting down...`, (11) create 5s timeout context, (12) call `server.Shutdown(ctx)`, (13) log `[claude-pulse] server stopped`. Use `slog` for all logging. Add to rootCmd via `init()`

**Checkpoint**: Full event pipeline works. `setup` + `serve` + WebSocket client demonstrates the core product. US2 auto-lifecycle is implicitly implemented by US1's hooks + this phase's server.

---

## Phase 5: User Story 4 - Device Connection Management (Priority: P2)

**Goal**: WebSocket connections are resilient — stale connections are detected via ping/pong, dead clients are evicted, and graceful close frames are sent on disconnect.

**Independent Test**: Connect a WebSocket client, verify it receives events. Kill the client abruptly, verify server logs disconnect and continues broadcasting to other clients. Reconnect and verify events resume.

### Implementation

- [x] T011 [US4] Add connection resilience to WebSocket in internal/server/websocket.go — add constants: `writeWait = 10s`, `pongWait = 60s`, `pingPeriod = 54s` (must be < pongWait), `maxMessageSize = 512`. In `readPump()`: set `conn.SetReadLimit(maxMessageSize)`, set `conn.SetReadDeadline(time.Now().Add(pongWait))`, set `conn.SetPongHandler` to reset read deadline. In `writePump()`: add `time.NewTicker(pingPeriod)`, on tick: set write deadline, send PingMessage, if error then return. On normal write: set `conn.SetWriteDeadline(time.Now().Add(writeWait))`. In `Hub.Shutdown()`: iterate all clients, send `websocket.CloseMessage` with `CloseNormalClosure` + "server shutting down", then close conn. Add `slog` logging: `websocket client connected`, `websocket client disconnected`, `broadcast to N client(s)`

**Checkpoint**: Device connections are production-resilient. Stale connections cleaned up automatically.

---

## Phase 6: User Story 5 - Simple Installation (Priority: P2)

**Goal**: Binary builds cleanly and CLI displays helpful usage information.

**Independent Test**: Run `go build -o claude-pulse .`, verify binary exists. Run `./claude-pulse` and verify help output. Run `./claude-pulse setup --help` and `./claude-pulse serve --help`.

### Implementation

- [x] T012 [US5] Verify build and CLI help output — run `go build -o claude-pulse .`, verify binary produces help matching CLI contract when run without args, verify `setup` and `serve` subcommands appear in help, verify `--help` flags work for each subcommand. Add version information: set `rootCmd.Version` to a build-time variable using `-ldflags`

**Checkpoint**: Binary is ready for distribution.

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases, logging consistency, and final hardening

- [x] T013 [P] Add consistent structured logging with slog across internal/server/ — ensure all log lines use `slog.Info`/`slog.Error` with key-value attrs: server start (port), event received (type), broadcast (type, client_count), client connect, client disconnect, shutdown. Use `[claude-pulse]` prefix via slog handler or group. Match log format from CLI contract
- [x] T014 [P] Handle edge cases across codebase — (1) internal/server/websocket.go: malformed WebSocket frames close the connection gracefully (readPump already handles via read error), (2) internal/server/handler.go: events with no connected clients are accepted (202) and broadcast to empty client set (no-op in Hub), (3) internal/hooks/setup.go: invalid settings.json returns descriptive error "error: .claude/settings.json contains invalid JSON" without overwriting the file

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (go.mod must exist)
- **US1 (Phase 3)**: Depends on Phase 2 (needs config for port)
- **US3+US2 (Phase 4)**: Depends on Phase 2 (needs config for port). Independent of US1.
- **US4 (Phase 5)**: Depends on Phase 4 (enhances existing websocket.go)
- **US5 (Phase 6)**: Depends on all implementation phases (needs buildable binary)
- **Polish (Phase 7)**: Depends on Phase 4 and Phase 5

### User Story Dependencies

- **US1 (P1)**: Independent — only needs config. Can be tested standalone (just writes files).
- **US3 (P1)**: Independent — only needs config. Can be tested with curl + websocat.
- **US2 (P1)**: Emerges from US1 + US3 combined. Auto-start hook (US1) + health endpoint + graceful shutdown (US3). Testable after both US1 and US3 complete.
- **US4 (P2)**: Enhances US3's WebSocket implementation. Must come after US3.
- **US5 (P2)**: Verifies the complete binary. Must come after all implementation.

### Within Each User Story

- Models/types before services/handlers
- Handlers before server wiring
- Server before CLI command
- Core implementation before resilience features

### Parallel Opportunities

**Phase 1**: T002 and T003 can run in parallel (after T001)
**Phase 3**: US1 can run in parallel with Phase 4 (US3+US2) — different packages
**Phase 4**: T007 and T008 can run in parallel (handler.go and websocket.go are independent)
**Phase 7**: T013 and T014 can run in parallel (different concerns)

---

## Parallel Example: Phase 4 (US3+US2)

```bash
# Launch these in parallel (different files, no dependencies):
Task: "Define Event types and POST /event handler in internal/server/handler.go"
Task: "Implement WebSocket Hub and Client in internal/server/websocket.go"

# Then sequentially (depends on both above):
Task: "Implement HTTP server in internal/server/server.go"
Task: "Implement serve cobra command in cmd/serve.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (T001-T003)
2. Complete Phase 2: Foundational (T004)
3. Complete Phase 3: US1 - Setup Command (T005-T006)
4. **STOP and VALIDATE**: Run `claude-pulse setup` in a test project, verify `.claude/settings.json`
5. This alone is useful — hooks are configured, ready for when server is built

### Core Product (US1 + US3 + US2)

6. Complete Phase 4: US3+US2 (T007-T010)
7. **STOP and VALIDATE**: Full pipeline works — setup a project, start server, connect WebSocket client, trigger events via curl
8. The product is functionally complete at this point

### Production-Ready (Add US4 + US5 + Polish)

9. Complete Phase 5: US4 (T011) — connection resilience
10. Complete Phase 6: US5 (T012) — verify binary and help
11. Complete Phase 7: Polish (T013-T014) — logging consistency and edge cases
12. **FINAL VALIDATION**: Run quickstart.md scenarios end-to-end

---

## Notes

- [P] tasks = different files, no dependencies
- [Story] label maps task to specific user story for traceability
- US2 (Auto Server Lifecycle) has no standalone tasks — it emerges from US1's hooks + US3's server
- Logging (FR-012) is addressed in T013 (polish phase) for consistency, but basic slog usage should be included as tasks are implemented
- Homebrew packaging (US5) is out of scope for initial implementation — US5 focuses on binary build and CLI help
- Commit after each task or logical group
- Stop at any checkpoint to validate independently
