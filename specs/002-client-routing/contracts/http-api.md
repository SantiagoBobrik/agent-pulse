# HTTP API Contract: Client Routing System

**Feature**: 002-client-routing
**Date**: 2026-02-25

## Server Endpoints (agent-pulse serve)

### GET /health

Health check endpoint (existing, unchanged).

**Response**: `200 OK` (empty body)

### POST /event

Receives events from Claude Code hooks. Dispatches to all registered HTTP clients via fan-out.

**Request**:
```
Content-Type: application/json
```

```json
{
  "type": "notification",
  "data": {
    "session_id": "abc-123",
    "message": "Claude needs your attention",
    "notification_type": "permission_prompt"
  }
}
```

**Validation**:
- `type` must be one of: `session_start`, `stop`, `notification`
- `data` is optional (but expected from hook subcommands)

**Response**: `202 Accepted` (empty body)

**Error Responses**:
- `400 Bad Request` — invalid JSON or unknown event type
- `405 Method Not Allowed` — non-POST request

---

## Client Delivery Contract (agent-pulse → registered clients)

When an event is dispatched, each subscribed client receives an HTTP POST.

### POST {client.url}

**Request**:
```
Content-Type: application/json
Authorization: Bearer {token}    ← only if auth configured
```

```json
{
  "type": "stop",
  "data": {
    "session_id": "abc-123",
    "message": "Task completed successfully"
  }
}
```

**Expected Client Response**: Any `2xx` status code = success.

**Timeout**: Per-client configured timeout (default: 2 seconds).

**Failure Handling**:
- Connection refused → logged as error, delivery skipped
- Timeout → logged as error, delivery skipped
- Non-2xx response → logged as error with status code, delivery skipped
- DNS resolution failure → logged as error, delivery skipped

No retries. Failures do not affect delivery to other clients.

---

## Event Payloads by Type

### session_start

```json
{
  "type": "session_start",
  "data": {
    "session_id": "abc-123"
  }
}
```

### stop

```json
{
  "type": "stop",
  "data": {
    "session_id": "abc-123",
    "message": "Here is the summary of what I did..."
  }
}
```

### notification

```json
{
  "type": "notification",
  "data": {
    "session_id": "abc-123",
    "message": "Claude needs your attention",
    "notification_type": "permission_prompt"
  }
}
```

---

## CLI Commands Contract

### agent-pulse hook --provider claude --event session_start

**Stdin**: JSON from Claude Code SessionStart hook
```json
{
  "session_id": "abc-123"
}
```

**Behavior**: Check server health → auto-start if needed → POST event to server

### agent-pulse hook --provider claude --event stop

**Stdin**: JSON from Claude Code Stop hook
```json
{
  "session_id": "abc-123",
  "last_assistant_message": "Done with the task"
}
```

**Behavior**: POST event to server with `message` from `last_assistant_message`

### agent-pulse hook --provider claude --event notification

**Stdin**: JSON from Claude Code Notification hook
```json
{
  "session_id": "abc-123",
  "message": "Claude needs attention",
  "notification_type": "permission_prompt"
}
```

**Behavior**: POST event to server with all fields in data

### agent-pulse client add

**Interactive mode** (no flags): Prompts for name, URL, port, timeout, events.
**Non-interactive mode** (flags): `--name`, `--url`, `--port`, `--timeout`, `--events`

**Exit codes**: 0 success, 1 validation error (duplicate name, invalid URL)

### agent-pulse client list

**Output format**:
```
NAME          URL                          TIMEOUT  EVENTS
escritorio    http://192.168.1.100         2s       all
slack-notif   https://hooks.slack.com/xxx  3s       notification
```

**Exit codes**: 0 always

### agent-pulse client remove \<name\>

**Exit codes**: 0 success, 1 client not found

### agent-pulse setup

**Behavior**: Writes Go binary hooks to `.claude/settings.json`

**Generated hooks format**:
```json
{
  "hooks": {
    "SessionStart": [{ "hooks": [{ "type": "command", "command": "agent-pulse hook --provider claude --event session_start" }] }],
    "Stop": [{ "hooks": [{ "type": "command", "command": "agent-pulse hook --provider claude --event stop" }] }],
    "Notification": [{ "hooks": [{ "type": "command", "command": "agent-pulse hook --provider claude --event notification" }] }]
  }
}
```
