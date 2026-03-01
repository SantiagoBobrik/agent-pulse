# Tasks: Client Routing System

**Input**: Design documents from `/specs/002-client-routing/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/http-api.md

**Tests**: Included — tech spec explicitly requires table-driven tests, httptest handler tests, and dispatcher tests.

**Organization**: Tasks grouped by user story. Each story is independently testable after its phase completes.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup

**Purpose**: Clean up WebSocket infrastructure and prepare for client routing

- [x] T001 Remove WebSocket hub — delete internal/server/websocket.go, remove gorilla/websocket from go.mod/go.sum, remove hub references from internal/server/server.go and internal/server/handler.go and cmd/serve.go

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core models and infrastructure that ALL user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

- [x] T002 [P] Extend Config struct with `BindAddress string` (default "127.0.0.1") and `Clients []Client` slice in internal/config/config.go — update Load() to parse new fields, update validation to accept bind_address
- [x] T003 [P] Create Client and Auth model structs with validation in internal/client/client.go — Client fields: Name, URL, Timeout (duration, default 2s), Events ([]string), Auth (*Auth). Auth fields: Type, Token. Add Validate() method (unique name check, URL parsing, event type validation). Add Accepts(eventType string) bool method for event filtering
- [x] T004 Create Dispatcher struct in internal/server/dispatcher.go — takes []Client from config, implements Dispatch(event Event) with goroutine fan-out, per-client http.Client with configured timeout, bearer auth header if configured, slog logging for success/failure/skip per client. Uses sync.WaitGroup to wait for all deliveries

**Checkpoint**: Foundation ready — Client model, Config extension, and Dispatcher engine are available

---

## Phase 3: User Story 1 — Register a New Client (Priority: P1) MVP

**Goal**: Users can register event destinations via interactive wizard or CLI flags

**Independent Test**: Run `agent-pulse client add`, complete the wizard, verify client appears in `~/.config/agent-pulse/config.yaml` with correct values

### Implementation for User Story 1

- [x] T005 [US1] Implement interactive wizard in internal/client/wizard.go — use bufio.Scanner for stdin prompts: name, URL/IP (prepend http:// if no scheme), port (default 80, append to URL), timeout (default 2s), event selection (all or pick from session_start/stop/notification). Return completed Client struct
- [x] T006 [US1] Implement `client` parent command and `client add` subcommand in cmd/client.go — flags: --name, --url, --port, --timeout, --events. If all required flags provided (name + url), skip wizard. Otherwise launch wizard from T005. Load config, validate uniqueness, append client, save config. Register parent command in cmd/root.go

### Tests for User Story 1

- [x] T007 [P] [US1] Table-driven tests for Client model validation in internal/client/client_test.go — test cases: valid client, empty name, invalid URL, timeout bounds, invalid event types, duplicate name detection, Accepts() filtering logic
- [x] T008 [P] [US1] Test config load/save with clients in internal/config/config_test.go — test cases: load config with clients array, save config preserving clients, default bind_address, empty clients list, round-trip serialization

**Checkpoint**: User Story 1 complete — clients can be registered and persisted

---

## Phase 4: User Story 2 — Distribute Events to Registered Clients (Priority: P1)

**Goal**: Events from Claude Code are delivered via HTTP POST to all subscribed clients in parallel

**Independent Test**: Register a test HTTP endpoint as a client, POST an event to /event, verify the endpoint receives the correct JSON payload

### Implementation for User Story 2

- [x] T009 [US2] Refactor Server struct to use Dispatcher instead of Hub in internal/server/server.go — remove Hub field, add Dispatcher field, pass Dispatcher to handler, update NewServer() constructor, update Shutdown() to no longer close hub
- [x] T010 [US2] Update event handler to dispatch via Dispatcher in internal/server/handler.go — on valid event, call dispatcher.Dispatch(event) instead of hub.broadcast, keep 202 response
- [x] T011 [US2] Update serve command to initialize Dispatcher in cmd/serve.go — load config, create Dispatcher with config.Clients, pass to NewServer(), remove hub goroutine
- [x] T012 [US2] Implement `hook` parent command with session_start, stop, notification subcommands in cmd/hook.go — each subcommand: read stdin JSON, parse typed fields per event type (see contracts/http-api.md), build Event{Type, Data}, HTTP POST to localhost:{port}/event. Register hook command in cmd/root.go

### Tests for User Story 2

- [x] T013 [P] [US2] Table-driven tests for Dispatcher in internal/server/dispatcher_test.go — test cases: fan-out to multiple clients, event filtering (client skips unsubscribed events), client timeout handling, client unreachable (connection refused), client returns non-2xx, empty events list means all events, bearer auth header sent when configured, one client failure doesn't block others
- [x] T014 [P] [US2] HTTP handler tests with httptest in internal/server/handler_test.go — test cases: valid event returns 202, invalid JSON returns 400, unknown event type returns 400, event dispatched to Dispatcher

**Checkpoint**: User Story 2 complete — events flow from hooks through server to registered clients

---

## Phase 5: User Story 3 — List and Manage Clients (Priority: P2)

**Goal**: Users can view all registered clients and remove clients by name

**Independent Test**: Register clients, run `agent-pulse client list` and verify table output, then run `agent-pulse client remove <name>` and verify removal

### Implementation for User Story 3

- [x] T015 [P] [US3] Add `client list` subcommand in cmd/client.go — load config, print formatted table: NAME / URL / TIMEOUT / EVENTS. No reachability checks (config display only)
- [x] T016 [P] [US3] Add `client remove` subcommand in cmd/client.go — takes client name as arg, load config, find and remove client by name, save config. Exit 1 if client not found with descriptive error

**Checkpoint**: User Story 3 complete — clients can be listed and removed

---

## Phase 6: User Story 4 — Setup Hook Registration (Priority: P2)

**Goal**: `agent-pulse setup` writes Go binary hook commands to `.claude/settings.json`

**Independent Test**: Run `agent-pulse setup` in a project directory, verify `.claude/settings.json` contains `agent-pulse hook --provider claude --event session_start/stop/notification` commands instead of curl/jq

### Implementation for User Story 4

- [x] T017 [US4] Update hook generation in internal/hooks/setup.go — replace curl/jq command strings with Go binary commands: `agent-pulse hook --provider claude --event session_start`, `agent-pulse hook --provider claude --event stop`, `agent-pulse hook --provider claude --event notification`. Keep merge logic that preserves existing non-agent-pulse hooks and prevents duplicate entries
- [x] T018 [P] [US4] Test hook generation output in internal/hooks/setup_test.go — test cases: fresh settings.json gets correct hook commands, existing hooks preserved, duplicate setup is idempotent, hook commands use Go binary format (no curl/jq)

**Checkpoint**: User Story 4 complete — setup writes Go-native hooks

---

## Phase 7: User Story 5 — Self-Healing Server Startup (Priority: P3)

**Goal**: `hook --provider claude --event session_start` auto-starts the server if it's not running

**Independent Test**: Ensure server is not running, run `agent-pulse hook --provider claude --event session_start` with stdin JSON, verify server starts and event is delivered

### Implementation for User Story 5

- [x] T019 [US5] Add health check and auto-start logic to hook command in cmd/hook.go (triggers on session_start event) — before dispatching: HTTP GET to localhost:{port}/health with 1s timeout. If healthy, dispatch normally. If unhealthy, exec `agent-pulse serve` as detached background process, poll /health up to 3s, then dispatch. Load port from config

**Checkpoint**: User Story 5 complete — server auto-starts on first session

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Edge cases and validation across all stories

- [x] T020 Handle edge cases across all packages — config missing/corrupted returns clear error in config.Load(), unknown event types logged as warning in handler, duplicate server detection via port binding error in server.go (already exists, verify behavior)
- [x] T021 Validate quickstart.md end-to-end flow — start server, register a test client, simulate event via curl, verify client receives payload

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies — start immediately
- **Foundational (Phase 2)**: Depends on Phase 1 (WebSocket removal clears the way)
- **US1 (Phase 3)**: Depends on Phase 2 (needs Client model and Config)
- **US2 (Phase 4)**: Depends on Phase 2 (needs Dispatcher and Client model)
- **US3 (Phase 5)**: Depends on Phase 2 (needs Config with clients). Benefits from US1 for test data but not strictly required
- **US4 (Phase 6)**: Depends on Phase 1 only (modifies hook generation, independent of client model)
- **US5 (Phase 7)**: Depends on US2 (hook subcommands must exist in cmd/hook.go)
- **Polish (Phase 8)**: Depends on all stories complete

### User Story Dependencies

```
Phase 1 (Setup)
    │
    ▼
Phase 2 (Foundational)
    │
    ├──────────────────┬──────────────────┐
    ▼                  ▼                  ▼
Phase 3 (US1)    Phase 4 (US2)    Phase 6 (US4)
    │                  │
    ▼                  ▼
Phase 5 (US3)    Phase 7 (US5)
    │                  │
    └──────────────────┘
              │
              ▼
      Phase 8 (Polish)
```

### Parallel Opportunities

**Within Phase 2 (Foundational)**:
- T002 (Config) and T003 (Client model) can run in parallel — different packages
- T004 (Dispatcher) depends on T003 (needs Client type)

**After Phase 2**:
- US1 (Phase 3) and US2 (Phase 4) can run in parallel — different files
- US4 (Phase 6) can run in parallel with US1 and US2 — modifies hooks/setup.go only

**Within US1 (Phase 3)**:
- T007 and T008 (tests) can run in parallel — different test files

**Within US2 (Phase 4)**:
- T013 and T014 (tests) can run in parallel — different test files

**Within US3 (Phase 5)**:
- T015 and T016 can run in parallel — different subcommands, same file but independent functions

---

## Parallel Example: After Phase 2

```bash
# These three phases can run in parallel:

# Stream A: US1 — Register Clients
Task: T005 "Implement wizard in internal/client/wizard.go"
Task: T006 "Implement client add command in cmd/client.go"
Task: T007 "Test client model in internal/client/client_test.go"
Task: T008 "Test config with clients in internal/config/config_test.go"

# Stream B: US2 — Event Distribution
Task: T009 "Refactor server to use dispatcher in internal/server/server.go"
Task: T010 "Update handler in internal/server/handler.go"
Task: T011 "Update serve command in cmd/serve.go"
Task: T012 "Implement hook subcommands in cmd/hook.go"

# Stream C: US4 — Hook Registration
Task: T017 "Update hook generation in internal/hooks/setup.go"
Task: T018 "Test hook generation in internal/hooks/setup_test.go"
```

---

## Implementation Strategy

### MVP First (User Story 1 + 2)

1. Complete Phase 1: Setup (remove WebSocket)
2. Complete Phase 2: Foundational (Config, Client model, Dispatcher)
3. Complete Phase 3: US1 (register clients)
4. Complete Phase 4: US2 (distribute events)
5. **STOP and VALIDATE**: Register a test endpoint, trigger an event, verify delivery
6. This is a functional MVP — events flow from Claude Code to registered clients

### Incremental Delivery

1. Phase 1 + 2 → Foundation ready
2. Add US1 → Can register clients (MVP building block)
3. Add US2 → Events flow end-to-end (MVP!)
4. Add US3 → Client management UX
5. Add US4 → Clean Go-native hooks replace curl/jq
6. Add US5 → Zero-friction server startup
7. Each story adds value without breaking previous stories

---

## Notes

- [P] tasks = different files, no dependencies on incomplete tasks
- [Story] label maps task to specific user story for traceability
- US1 and US2 are both P1 — together they form the MVP
- US3 and US4 are P2 — management and setup improvements
- US5 is P3 — quality-of-life auto-start
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
