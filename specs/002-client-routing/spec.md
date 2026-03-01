# Feature Specification: Client Routing System

**Feature Branch**: `002-client-routing`
**Created**: 2026-02-25
**Status**: Draft
**Input**: User description: "agent-pulse client routing system — distribute Claude Code lifecycle events to registered HTTP clients"
**Extends**: Feature 001 (Event Bridge) — builds client routing on top of the existing server, hooks, and CLI infrastructure.

## Clarifications

### Session 2026-02-25

- Q: Are clients global (all projects) or per-project? → A: Global — all registered clients receive events from every project that has run `setup`.
- Q: What fields does the event HTTP POST payload contain? → A: JSON with `type` (string) and `data` (object with event-specific fields like `session_id`, `message`, `notification_type`).
- Q: Is the server endpoint restricted to localhost? → A: Configurable bind address (default `127.0.0.1`, can bind to `0.0.0.0`), no authentication on the event ingestion endpoint.
- Q: Should `client add` support non-interactive mode? → A: Both — interactive wizard when no flags provided, non-interactive when all required flags are given.
- Q: Relationship to feature 001 (event bridge)? → A: Extend — build client routing into the existing event bridge codebase.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Register a New Client (Priority: P1)

A developer wants to register a new destination (e.g., an ESP32 on their desk, a Slack webhook, or a local script) to receive Claude Code events. They run a guided wizard that asks for the client's name, URL, port, timeout, event filters, and authentication. Once completed, the client is saved to the global configuration and immediately eligible to receive events from any project where `setup` has been run. Alternatively, they can provide all required fields as flags for non-interactive/scripted usage.

**Why this priority**: Without the ability to register clients, no events can be delivered anywhere. This is the foundational capability that enables the entire system.

**Independent Test**: Can be fully tested by running `agent-pulse client add`, completing the wizard, and verifying the client appears in the configuration file with correct settings.

**Acceptance Scenarios**:

1. **Given** no clients are registered, **When** the user runs `agent-pulse client add` and provides name "escritorio", URL "192.168.1.100", port 80, timeout 2s, events "all", auth "none", **Then** the client "escritorio" is saved to the global config file with all provided values and a success message is displayed.
2. **Given** a client named "escritorio" already exists, **When** the user runs `agent-pulse client add` with name "escritorio", **Then** the system rejects the registration and informs the user that the name is already taken.
3. **Given** the wizard is running, **When** the user provides an invalid URL (e.g., empty string or malformed address), **Then** the system shows an error and re-prompts for a valid URL.
4. **Given** the wizard is running, **When** the user selects specific events to receive, **Then** only the selected events are saved in the client's filter configuration.
5. **Given** no flags are provided, **When** the user runs `agent-pulse client add`, **Then** the interactive wizard starts and prompts for each field sequentially.
6. **Given** all required flags are provided (e.g., `--name slack --url https://hooks.example.com --events notification`), **When** the user runs `agent-pulse client add`, **Then** the client is registered without interactive prompts.

---

### User Story 2 - Distribute Events to Registered Clients (Priority: P1)

When Claude Code triggers a lifecycle event (session_start, stop, or notification), the running server receives it and distributes it to all globally registered clients whose event filters include that event type. Each client is contacted via HTTP POST with a JSON payload containing `type` and an `data` object with event-specific fields (e.g., `session_id`, `message`, `notification_type`). Delivery to each client is independent — one client's failure does not affect others.

**Why this priority**: This is the core value proposition of agent-pulse. Without event distribution, registering clients has no purpose.

**Independent Test**: Can be tested by registering a test HTTP endpoint as a client, triggering a Claude Code event, and verifying the endpoint receives the correct payload.

**Acceptance Scenarios**:

1. **Given** two clients are registered (both subscribing to "stop"), **When** a "stop" event occurs, **Then** both clients receive an HTTP POST with a JSON body containing `type` and `data` (with `session_id` and `message`).
2. **Given** client A subscribes to "stop" only and client B subscribes to all events, **When** a "session_start" event occurs, **Then** only client B receives the event.
3. **Given** client A is unreachable (network error), **When** an event is dispatched, **Then** client B still receives the event successfully and client A's failure is logged.
4. **Given** a client has a 2-second timeout configured, **When** the client takes longer than 2 seconds to respond, **Then** the delivery is marked as failed (timeout) and the server moves on without blocking other deliveries.

---

### User Story 3 - List and Manage Registered Clients (Priority: P2)

A developer wants to see which clients are registered. They run `agent-pulse client list` to see a table of all clients with their configuration. They can also remove a client with `agent-pulse client remove <name>`.

**Why this priority**: Management commands are important for usability but not strictly required for event delivery to work.

**Independent Test**: Can be tested by registering clients, running `client list` to verify the output, then running `client remove` and confirming the client is gone.

**Acceptance Scenarios**:

1. **Given** three clients are registered, **When** the user runs `agent-pulse client list`, **Then** all three clients are displayed with their name, URL, timeout, and event filter.
2. **Given** client "escritorio" is registered, **When** the user runs `agent-pulse client remove escritorio`, **Then** the client is removed from the config and a confirmation message is displayed.
3. **Given** no client named "phantom" exists, **When** the user runs `agent-pulse client remove phantom`, **Then** the system shows an error message indicating the client was not found.
4. **Given** a client's endpoint is down, **When** the user runs `agent-pulse client list`, **Then** the client is still displayed (list shows configuration only, no reachability check).

---

### User Story 4 - Setup Hook Registration (Priority: P2)

A developer runs `agent-pulse setup` once per project to register the necessary hooks in `.claude/settings.json`. After setup, Claude Code automatically sends lifecycle events to the agent-pulse server without any further user action. Clients are global — once registered, they receive events from all projects where setup has been run.

**Why this priority**: Setup is a one-time operation that bridges Claude Code to agent-pulse. Important but only needs to work once per project.

**Independent Test**: Can be tested by running `agent-pulse setup` in a project directory and verifying that `.claude/settings.json` contains the correct hook entries for session_start, stop, and notification events.

**Acceptance Scenarios**:

1. **Given** a project directory without existing hooks, **When** the user runs `agent-pulse setup`, **Then** `.claude/settings.json` is created/updated with hooks for session_start, stop, and notification events.
2. **Given** a project already has other hooks in `.claude/settings.json`, **When** the user runs `agent-pulse setup`, **Then** the agent-pulse hooks are added without overwriting existing hooks.
3. **Given** agent-pulse hooks are already registered, **When** the user runs `agent-pulse setup` again, **Then** the hooks are not duplicated and a message indicates setup is already complete.

---

### User Story 5 - Self-Healing Server Startup (Priority: P3)

When a new Claude Code session starts, the session_start hook checks whether the agent-pulse server is already running. If not, it starts the server automatically in the background. The user never needs to manually run `agent-pulse serve`.

**Why this priority**: This is a quality-of-life improvement. The server could be started manually; auto-start just removes friction.

**Independent Test**: Can be tested by ensuring the server is not running, then triggering a session_start hook and verifying the server starts automatically and begins accepting events.

**Acceptance Scenarios**:

1. **Given** the agent-pulse server is not running, **When** a session_start hook fires, **Then** the server is started in the background and the hook event is delivered to registered clients.
2. **Given** the agent-pulse server is already running, **When** a session_start hook fires, **Then** the server is not restarted and the event is delivered normally.
3. **Given** the server process crashed, **When** the next session_start hook fires, **Then** a new server process is started and events resume delivery.

---

### Edge Cases

- What happens when the config file is missing or corrupted? The system should report a clear error and not crash.
- What happens when a client's URL resolves but the endpoint returns a non-2xx HTTP status? The delivery should be logged as failed with the status code.
- What happens when all registered clients are unreachable? Events should be logged locally and the server should continue running.
- What happens when the server receives an event type it doesn't recognize? The event should be ignored with a warning log.
- What happens when two `agent-pulse serve` processes are started simultaneously? The second process should detect the first (e.g., via port conflict or PID file) and exit gracefully.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide an interactive wizard (`agent-pulse client add`) that collects client name, URL/IP, port, timeout, event filter, and authentication configuration.
- **FR-002**: System MUST support non-interactive client registration via flags (e.g., `--name`, `--url`, `--port`, `--timeout`, `--events`) when all required fields are provided. Auth is configured by editing the YAML config directly.
- **FR-003**: System MUST validate client names are unique within the configuration.
- **FR-004**: System MUST validate client URLs/IPs are well-formed before saving.
- **FR-005**: System MUST support three event types: `session_start`, `stop`, and `notification`.
- **FR-006**: System MUST allow each client to subscribe to a subset of events or all events (default: all).
- **FR-007**: System MUST deliver events via HTTP POST with a JSON payload containing `type` (string) and `data` (object with event-specific fields such as `session_id`, `message`, `notification_type`).
- **FR-008**: System MUST respect per-client timeout configuration when delivering events (default: 2 seconds).
- **FR-009**: System MUST NOT let one client's delivery failure affect delivery to other clients.
- **FR-010**: System MUST provide `agent-pulse client list` showing all registered clients with their configuration (name, URL, timeout, events).
- **FR-011**: System MUST provide `agent-pulse client remove <name>` to delete a client from the configuration.
- **FR-012**: System MUST provide `agent-pulse setup` to register hooks in `.claude/settings.json` without overwriting existing hooks.
- **FR-013**: System MUST provide `agent-pulse serve` to run a background server that receives events and distributes them to clients.
- **FR-014**: System MUST support bearer token authentication per client for secured endpoints.
- **FR-015**: System MUST auto-start the server when a session_start hook fires and the server is not running.
- **FR-016**: System MUST detect duplicate server processes and prevent multiple instances from running simultaneously.
- **FR-017**: System MUST persist client configuration globally in `~/.config/agent-pulse/config.yaml`. Clients receive events from all projects where `setup` has been run.
- **FR-018**: System MUST log delivery failures (timeouts, connection errors, non-2xx responses) with sufficient detail for debugging.
- **FR-019**: System MUST allow configuring the server bind address (default: `127.0.0.1`) to support both localhost-only and network-accessible deployments.
- **FR-020**: System MUST extend the existing event bridge infrastructure from feature 001 rather than creating a parallel server or CLI.

### Key Entities

- **Client**: A registered event destination with a name (unique identifier), URL/IP, port, timeout duration, list of subscribed events, and optional bearer token. Clients are global — not scoped to a specific project.
- **Event**: A lifecycle signal from Claude Code delivered as a JSON payload with `type` (session_start | stop | notification) and `data` (object containing event-specific fields like `session_id`, `message`, `notification_type`).
- **Configuration**: The persistent global store of all registered clients and server settings, stored as YAML at `~/.config/agent-pulse/config.yaml`.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can register a new client in under 60 seconds using the interactive wizard.
- **SC-002**: Events are delivered to all reachable, subscribed clients within 1 second of the event occurring (excluding client response time).
- **SC-003**: One client's failure (timeout, unreachable, error) does not delay delivery to other clients by more than 100 milliseconds.
- **SC-004**: Users can set up a project for event tracking (`agent-pulse setup`) in a single command with no manual file editing.
- **SC-005**: The server auto-starts on the first Claude Code session without any user intervention after initial setup.
- **SC-006**: Users can list all registered clients and their configuration in a single command.
- **SC-007**: The system supports at least 20 simultaneously registered clients without performance degradation.

## Assumptions

- This feature extends the existing event bridge from feature 001, reusing its server, CLI, and hook infrastructure.
- The user has Go 1.24+ installed (consistent with the existing project stack).
- Client endpoints are HTTP-based and accept POST requests with JSON payloads.
- The configuration file location (`~/.config/agent-pulse/config.yaml`) is accessible and writable by the user.
- Bearer token is the only supported authentication mechanism (covers most use cases; additional auth methods can be added later).
- Event delivery is fire-and-forget (no retry mechanism) — if a client is unreachable, the event is logged and dropped.
- The server listens on a configurable local port (default: 8080) with a configurable bind address (default: `127.0.0.1`).
- The event ingestion endpoint does not require authentication (events are accepted from any source that can reach the bind address).
