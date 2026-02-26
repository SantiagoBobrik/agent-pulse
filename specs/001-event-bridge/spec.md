# Feature Specification: Event Bridge

**Feature Branch**: `001-event-bridge`
**Created**: 2026-02-25
**Status**: Draft
**Input**: User description: "claude-pulse — a bridge between Claude Code and physical hardware that forwards lifecycle events to connected devices via WebSocket"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - One-Time Project Setup (Priority: P1)

A developer wants to connect their Claude Code project to claude-pulse so that lifecycle events are automatically forwarded to their devices. They run a single setup command in their project directory. The command configures Claude Code to send events to claude-pulse without requiring any manual file editing or ongoing maintenance.

**Why this priority**: Without setup, no events can flow. This is the entry point for the entire system and must work flawlessly on first use.

**Independent Test**: Can be fully tested by running the setup command in a project directory and verifying that the correct hook configuration is written to the project's settings file.

**Acceptance Scenarios**:

1. **Given** a project directory with no existing claude-pulse configuration, **When** the user runs the setup command, **Then** the project's Claude Code settings are updated with the correct event hooks for SessionStart, Stop, and Notification.
2. **Given** a project directory with existing Claude Code settings (unrelated hooks or preferences), **When** the user runs the setup command, **Then** the existing settings are preserved and claude-pulse hooks are merged in without overwriting anything.
3. **Given** a project directory that has already been set up with claude-pulse, **When** the user runs the setup command again, **Then** the hooks are updated to the latest version without creating duplicates.

---

### User Story 2 - Automatic Server Lifecycle (Priority: P1)

A developer starts a Claude Code session in a project that has been set up with claude-pulse. The claude-pulse server starts automatically in the background without any manual intervention. When the developer uses Claude Code normally, events flow to connected devices. The developer never has to think about starting or managing the server.

**Why this priority**: The zero-maintenance promise is central to the product. If users have to manually start the server, adoption drops significantly.

**Independent Test**: Can be fully tested by starting a Claude Code session in a configured project and verifying the server starts automatically and begins accepting connections.

**Acceptance Scenarios**:

1. **Given** a configured project with no claude-pulse server running, **When** a Claude Code session starts, **Then** the server is automatically started in the background and a SessionStart event is sent.
2. **Given** a configured project with the claude-pulse server already running, **When** a Claude Code session starts, **Then** the existing server is reused (not duplicated) and a SessionStart event is sent.
3. **Given** the server was started automatically, **When** the server process receives a termination signal, **Then** it shuts down gracefully, closing all active connections cleanly.

---

### User Story 3 - Real-Time Event Forwarding (Priority: P1)

A developer is using Claude Code while away from their desk. Their connected device (e.g., an ESP32 on the local network) receives real-time updates about what Claude is doing — whether it started a session, finished responding, or needs attention. The developer can glance at their device to know the current state without returning to their terminal.

**Why this priority**: This is the core value proposition. Without event forwarding to devices, the product has no purpose.

**Independent Test**: Can be fully tested by connecting a WebSocket client to the server, triggering each event type, and verifying the client receives the correct event data in real-time.

**Acceptance Scenarios**:

1. **Given** a connected device listening via WebSocket, **When** Claude Code starts a session, **Then** the device receives a SessionStart event immediately.
2. **Given** a connected device listening via WebSocket, **When** Claude Code finishes responding and is waiting for input, **Then** the device receives a Stop event.
3. **Given** a connected device listening via WebSocket, **When** Claude Code needs user attention (permission prompt, idle, etc.), **Then** the device receives a Notification event with the relevant message and notification type.
4. **Given** multiple devices connected via WebSocket, **When** any event occurs, **Then** all connected devices receive the event simultaneously.

---

### User Story 4 - Device Connection Management (Priority: P2)

A device connects to the claude-pulse server over WebSocket to receive events. If the device disconnects (due to network issues, power cycle, etc.), it can reconnect and resume receiving events. The server tracks connected clients and handles connections/disconnections gracefully.

**Why this priority**: Device reliability is important but secondary to the core event flow. The system must handle real-world network conditions.

**Independent Test**: Can be fully tested by connecting and disconnecting WebSocket clients and verifying the server tracks them correctly and continues broadcasting to remaining clients.

**Acceptance Scenarios**:

1. **Given** the server is running, **When** a device connects via WebSocket, **Then** the device is added to the client pool and begins receiving events.
2. **Given** a device is connected, **When** the device disconnects, **Then** the device is removed from the client pool and no errors occur during subsequent broadcasts.
3. **Given** a device previously disconnected, **When** the device reconnects, **Then** it is re-added to the client pool and resumes receiving events.

---

### User Story 5 - Simple Installation (Priority: P2)

A developer installs claude-pulse via their package manager. After installation, they have access to the CLI commands and can immediately set up their first project.

**Why this priority**: Installation is a one-time action and standard for CLI tools. Important for adoption but not core functionality.

**Independent Test**: Can be fully tested by installing the package and verifying the CLI binary is available and responds to help commands.

**Acceptance Scenarios**:

1. **Given** the user has a package manager available, **When** they install claude-pulse, **Then** the CLI binary is available in their PATH and responds to commands.
2. **Given** claude-pulse is installed, **When** the user runs it without arguments, **Then** helpful usage information is displayed.

---

### Edge Cases

- When the server port is already in use by another application, the server MUST fail with a clear error message identifying the blocked port and guiding the user to change it in the configuration file.
- What happens when a device sends malformed WebSocket frames?
- What happens when an event is received but no devices are connected?
- What happens when the network is unavailable and a device cannot reach the server?
- What happens when the settings file has invalid JSON before setup runs?
- When multiple Claude Code sessions run simultaneously in different projects, they all share the same server instance. Events from any project are broadcast to all connected devices.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST provide a setup command that configures a project's Claude Code hooks to send lifecycle events to the server.
- **FR-002**: System MUST merge hook configuration into existing project settings without overwriting unrelated settings.
- **FR-003**: System MUST provide a server that receives events via HTTP and forwards them to connected devices via WebSocket.
- **FR-004**: System MUST handle three event types: SessionStart, Stop, and Notification.
- **FR-005**: System MUST automatically start the server when a Claude Code session begins if the server is not already running.
- **FR-006**: System MUST detect if the server is already running to avoid starting duplicate instances.
- **FR-007**: System MUST broadcast received events to all connected WebSocket clients simultaneously.
- **FR-008**: System MUST track connected clients and handle connections, disconnections, and reconnections gracefully.
- **FR-009**: System MUST provide a health check endpoint that reports whether the server is running.
- **FR-010**: System MUST shut down gracefully when receiving a termination signal, closing all active WebSocket connections before exiting.
- **FR-011**: System MUST support configurable server port via a configuration file.
- **FR-012**: System MUST log all key events (server start, client connect/disconnect, events received, broadcasts sent) in a structured format.
- **FR-013**: System MUST accept an optional `extra` field with event-type-specific data (e.g., `message` and `notification_type` for Notification events). Stop and SessionStart events carry no extra data.

### Key Entities

- **Event**: A lifecycle occurrence in Claude Code. Has a type (SessionStart, Stop, Notification) and optional extra data that varies by type. Events flow from Claude Code hooks to the server and then to connected devices.
- **Client**: A device connected to the server via WebSocket. Tracked by session ID. Can connect, disconnect, and reconnect. Receives broadcast events.
- **Hook**: A shell command configured in a Claude Code project that fires on a specific lifecycle event. Written by the setup command. Responsible for health-checking the server, auto-starting it, and sending event data.
- **Configuration**: User-specific settings including server port and device connection details. Stored in a standard configuration directory.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: A new user can go from installation to receiving their first event on a connected device in under 5 minutes.
- **SC-002**: Events are delivered to connected devices within 1 second of occurring in Claude Code.
- **SC-003**: The server starts automatically on the first Claude Code session with zero manual intervention from the user.
- **SC-004**: The setup command completes in under 3 seconds and does not corrupt existing project settings.
- **SC-005**: The system supports at least 10 simultaneous WebSocket client connections without degradation.
- **SC-006**: The server remains stable during continuous operation across multiple Claude Code sessions without requiring restarts.
- **SC-007**: Device disconnections and reconnections are handled without errors or event loss to other connected clients.

## Clarifications

### Session 2026-02-25

- Q: Should there be a maximum size for event payloads forwarded to devices? → A: Remove `last_assistant_message` from Stop events entirely. Stop events carry only the event type with no extra data.
- Q: What should happen when the configured port is already in use? → A: Fail with a clear error message and guidance to change the port in config.
- Q: Should there be one shared server or one per project? → A: One shared server for all projects; events from any project are broadcast to all connected devices.

## Assumptions

- The user has Claude Code installed and is using it in a project with a `.claude` directory (or one will be created).
- The user's local network allows WebSocket connections between the server and connected devices.
- The server runs on the same machine as Claude Code (localhost communication for hooks).
- Homebrew is the primary distribution channel for macOS users; other package managers may be supported later.
- Events do not need to be persisted or queued — if no devices are connected, events are discarded.
- The configuration file uses a standard OS-specific configuration directory (e.g., `~/.config/` on macOS/Linux).
