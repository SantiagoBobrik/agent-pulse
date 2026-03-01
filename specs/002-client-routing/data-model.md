# Data Model: Client Routing System

**Feature**: 002-client-routing
**Date**: 2026-02-25

## Entities

### Config

The root configuration object. Extends the existing `Config` struct from feature 001.

| Field        | Type       | Required | Default       | Description                          |
| ------------ | ---------- | -------- | ------------- | ------------------------------------ |
| port         | int        | no       | 8080          | Server listen port                   |
| bind_address | string     | no       | "127.0.0.1"   | Server bind address                  |
| clients      | []Client   | no       | []            | Registered event destinations        |

### Client

A registered event destination.

| Field   | Type     | Required | Default | Validation                                                  |
| ------- | -------- | -------- | ------- | ----------------------------------------------------------- |
| name    | string   | yes      | —       | Unique, non-empty, alphanumeric + hyphens + underscores     |
| url     | string   | yes      | —       | Valid URL with scheme (http:// prepended if missing)         |
| timeout | duration | no       | "2s"    | Go duration string, must be > 0 and ≤ 30s                   |
| events  | []string | no       | []      | Empty = all events. Valid values: session_start, stop, notification |
| auth    | *Auth    | no       | nil     | Optional authentication config                              |

**Uniqueness**: Client name must be unique within the clients list.

**Event filter logic**: If `events` is empty or omitted, the client receives all event types. If populated, only listed event types are delivered.

### Auth

Optional authentication for client endpoints.

| Field | Type   | Required | Default | Description                                |
| ----- | ------ | -------- | ------- | ------------------------------------------ |
| type  | string | yes      | —       | Auth type. Only "bearer" supported.        |
| token | string | yes      | —       | Bearer token value. Supports `${ENV_VAR}` syntax for environment variable expansion. |

### Event

A lifecycle signal from Claude Code. Uses the existing `Event` struct.

| Field | Type            | Required | Description                                    |
| ----- | --------------- | -------- | ---------------------------------------------- |
| type  | string          | yes      | One of: session_start, stop, notification      |
| data  | json.RawMessage | no       | Event-specific data (see Data Fields below)    |

#### Data Fields by Event Type

| Event Type     | Data Fields                                          |
| -------------- | ----------------------------------------------------- |
| session_start  | `session_id` (string)                                 |
| stop           | `session_id` (string), `message` (string)             |
| notification   | `session_id` (string), `message` (string), `notification_type` (string) |

## YAML Configuration Example

```yaml
port: 8080
bind_address: "127.0.0.1"
clients:
  - name: escritorio
    url: http://192.168.1.100
    timeout: 2s
    events: ["session_start", "stop", "notification"]

  - name: slack-notif
    url: https://hooks.slack.com/xxx
    timeout: 3s
    events: ["notification"]
    auth:
      type: bearer
      token: ${SLACK_TOKEN}
```

## State Transitions

### Server Lifecycle

```
Not Running → (hook --provider claude --event session_start checks /health, fails) → Starting → Running
Running → (SIGINT/SIGTERM) → Shutting Down → Not Running
Running → (port conflict on second instance) → Error (exit)
```

### Event Dispatch Flow

```
Event Received → Validate Type → Fan-out to Matching Clients → Log Results
                                  ├── Client A: OK (200)
                                  ├── Client B: Timeout
                                  └── Client C: Skipped (filtered)
```

## Relationships

```
Config 1──* Client
Client *──1 Auth (optional)
Server 1──1 Config
Server 1──1 Dispatcher
Dispatcher *──* Client (reads from config)
Event ──→ Dispatcher ──→ Client(s)
```
