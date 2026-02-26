# HTTP API Contract

**Date**: 2026-02-25
**Branch**: `001-event-bridge`

## Endpoints

### GET /health

Health check endpoint. Used by SessionStart hook to detect if server is running.

**Response**:
- `200 OK` — server is running (body: empty or `.`)

### POST /event

Receives lifecycle events from Claude Code hooks.

**Request**:
- Content-Type: `application/json`

**Request Body**:
```json
{
  "type": "session_start | stop | notification",
  "extra": {}
}
```

**Event schemas**:

#### session_start
```json
{
  "type": "session_start"
}
```

#### stop
```json
{
  "type": "stop"
}
```

#### notification
```json
{
  "type": "notification",
  "extra": {
    "message": "Claude needs your attention",
    "notification_type": "permission_prompt"
  }
}
```

Valid `notification_type` values: `permission_prompt`, `idle_prompt`, `auth_success`, `elicitation_dialog`

**Responses**:
- `202 Accepted` — event received and will be broadcast
- `400 Bad Request` — invalid JSON or unknown event type

### GET /ws (WebSocket Upgrade)

WebSocket endpoint for devices to connect and receive events.

**Upgrade**: Standard WebSocket handshake (`Upgrade: websocket`)

**Server → Client messages**: JSON-encoded events matching the POST /event schema:

```json
{
  "type": "notification",
  "extra": {
    "message": "Claude needs your attention",
    "notification_type": "permission_prompt"
  }
}
```

**Client → Server**: No client-to-server messages expected. Server reads only to detect disconnection and handle ping/pong.

**Connection lifecycle**:
1. Client connects via WebSocket upgrade
2. Server registers client in pool
3. Server sends events as they arrive (JSON text frames)
4. On disconnect or error, server removes client from pool
5. Client may reconnect at any time

**Ping/Pong**: Server sends periodic pings. Clients must respond with pong within the deadline or the connection is closed.
