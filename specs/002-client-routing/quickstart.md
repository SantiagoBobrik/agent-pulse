# Quickstart: Client Routing System

**Feature**: 002-client-routing

## Prerequisites

- Go 1.24+ installed
- agent-pulse binary built and in PATH
- A target endpoint to receive events (ESP32, webhook URL, local HTTP server)

## Setup (one time per project)

```bash
# Register hooks in the current project
agent-pulse setup
```

This writes hook entries to `.claude/settings.json` so Claude Code automatically sends lifecycle events.

## Register a Client

### Interactive wizard

```bash
agent-pulse client add
```

Follow the prompts to enter name, URL, port, timeout, and event filter.

### Non-interactive (scripting)

```bash
# Register an ESP32 device
agent-pulse client add --name escritorio --url http://192.168.1.100

# Register a Slack webhook for notifications only
agent-pulse client add --name slack --url https://hooks.slack.com/xxx --events notification --timeout 3s
```

## Manage Clients

```bash
# List all registered clients
agent-pulse client list

# Remove a client
agent-pulse client remove escritorio
```

## How It Works

1. Claude Code triggers a lifecycle event (session_start, stop, notification)
2. The hook runs `agent-pulse hook --provider <provider> --event <event>`, which reads stdin and POSTs to the server
3. If the server isn't running, `hook --event session_start` auto-starts it
4. The server fans out the event to all registered clients matching the event filter
5. Each client receives an HTTP POST with the event payload

## Event Payload (what clients receive)

```json
{
  "type": "stop",
  "data": {
    "session_id": "abc-123",
    "message": "Task completed"
  }
}
```

## Configuration

Config lives at `~/.config/agent-pulse/config.yaml`:

```yaml
port: 8080
bind_address: "127.0.0.1"
clients:
  - name: escritorio
    url: http://192.168.1.100
    timeout: 2s
    events: []  # empty = all events
  - name: slack-notif
    url: https://hooks.slack.com/xxx
    timeout: 3s
    events: ["notification"]
    auth:
      type: bearer
      token: ${SLACK_TOKEN}
```

## Testing the Setup

```bash
# Start server manually (or let hook auto-start it)
agent-pulse serve

# In another terminal, simulate an event
curl -X POST http://localhost:8080/event \
  -H 'Content-Type: application/json' \
  -d '{"type":"stop","data":{"session_id":"test","message":"hello"}}'
```

All registered clients should receive the event.
