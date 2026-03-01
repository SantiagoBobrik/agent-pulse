# Quickstart: Event Bridge

**Date**: 2026-02-25
**Branch**: `001-event-bridge`

## Prerequisites

- Go 1.24+
- Claude Code installed

## Build & Run

```bash
# Build
go build -o agent-pulse .

# Setup in a project (run from project root)
./agent-pulse setup

# Start server manually (normally auto-started by hooks)
./agent-pulse serve
```

## Verify Setup

After running `setup`, check `.claude/settings.json` in your project:

```bash
cat .claude/settings.json | jq '.hooks'
```

You should see `SessionStart`, `Stop`, and `Notification` hook entries.

## Test Event Flow

```bash
# Terminal 1: Start server
./agent-pulse serve

# Terminal 2: Connect a WebSocket client
websocat ws://localhost:8080/ws

# Terminal 3: Send a test event
curl -X POST http://localhost:8080/event \
  -H 'Content-Type: application/json' \
  -d '{"type":"session_start"}'
```

The WebSocket client in Terminal 2 should receive the event.

## Test Notification Event

```bash
curl -X POST http://localhost:8080/event \
  -H 'Content-Type: application/json' \
  -d '{"type":"notification","data":{"message":"Claude needs attention","notification_type":"permission_prompt"}}'
```

## Health Check

```bash
curl http://localhost:8080/health
# Returns 200 OK if server is running
```

## Configuration

Default config location: `~/.config/agent-pulse/config.yaml`

```yaml
port: 8080
```
