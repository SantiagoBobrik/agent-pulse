# Data Model: Event Bridge

**Date**: 2026-02-25
**Branch**: `001-event-bridge`

## Entities

### Event

An event received from Claude Code hooks and broadcast to connected devices.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| type  | string (enum) | yes | One of: `session_start`, `stop`, `notification` |
| data  | object | no | Event-type-specific data. Absent for `session_start` and `stop`. |

**Data fields by event type**:

| Event Type | Data Field | Type | Description |
|------------|-------------|------|-------------|
| `notification` | `message` | string | Human-readable notification message |
| `notification` | `notification_type` | string | One of: `permission_prompt`, `idle_prompt`, `auth_success`, `elicitation_dialog` |

**Validation rules**:
- `type` must be one of the three allowed values
- `data` is ignored for `session_start` and `stop`
- For `notification`, at least `message` must be present in `data`

### Client

A WebSocket connection tracked by the server's client pool.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| conn | WebSocket connection | yes | The underlying WebSocket connection |
| send | buffered channel | yes | Outbound message queue for this client |

**Lifecycle**:
- **Connected**: Client upgrades HTTP → WebSocket, registered in Hub
- **Active**: Receives broadcast events via `send` channel
- **Disconnected**: Removed from Hub on read error, close frame, or send buffer full

**State transitions**:
```
[New Connection] → Connected → Active ⇄ Receiving Events
                                  ↓
                            Disconnected → [Removed from Hub]
                                  ↓
                         [New Connection] → Connected (reconnect)
```

### Configuration

User settings for the agent-pulse server.

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| port | integer | no | 8080 | Server listen port |

**Storage**: YAML file at `~/.config/agent-pulse/config.yaml`

**Validation rules**:
- `port` must be between 1024 and 65535
- If config file doesn't exist, use defaults

### Hook Configuration

Written by the `setup` command into `.claude/settings.json`.

| Field | Type | Description |
|-------|------|-------------|
| hooks | object | Top-level key in settings.json |
| hooks.SessionStart | array | Hook entries for session start events |
| hooks.Stop | array | Hook entries for stop events |
| hooks.Notification | array | Hook entries for notification events |

Each hook entry follows the Claude Code schema:
```json
{
  "hooks": [{
    "type": "command",
    "command": "<shell command>"
  }]
}
```

## Relationships

```
Claude Code Hooks  ──POST /event──▶  Server  ──broadcast──▶  Client Pool
                                       │                        │
                                   reads config            0..N Clients
                                       │
                                  config.yaml
```

- Server has 0..N Clients (one-to-many)
- Server reads 1 Configuration (one-to-one)
- Events flow unidirectionally: Hooks → Server → Clients
- No persistence — events are fire-and-forget
