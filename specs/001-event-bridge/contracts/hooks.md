# Hook Configuration Contract

**Date**: 2026-02-25
**Branch**: `001-event-bridge`

## Settings Written by `agent-pulse setup`

The setup command writes the following structure into `.claude/settings.json` under the `hooks` key:

```json
{
  "hooks": {
    "SessionStart": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "curl -sf http://localhost:8080/health > /dev/null 2>&1 || (agent-pulse serve > /dev/null 2>&1 &); sleep 1; curl -sf -X POST http://localhost:8080/event -H 'Content-Type: application/json' -d '{\"type\":\"session_start\"}'"
          }
        ]
      }
    ],
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "curl -sf -X POST http://localhost:8080/event -H 'Content-Type: application/json' -d '{\"type\":\"stop\"}'"
          }
        ]
      }
    ],
    "Notification": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "jq -c '{type:\"notification\",data:{message:.message,notification_type:.notification_type}}' | curl -sf -X POST http://localhost:8080/event -H 'Content-Type: application/json' -d @-"
          }
        ]
      }
    ]
  }
}
```

## Hook Behavior

### SessionStart Hook
1. Check if server is running (`curl /health`)
2. If not running, start server in background (`agent-pulse serve &`)
3. Wait briefly for server to be ready
4. Send `session_start` event

### Stop Hook
1. Send `stop` event (no data data)
2. Does not read stdin (ignores `last_assistant_message`)

### Notification Hook
1. Read stdin JSON from Claude Code
2. Extract `message` and `notification_type` via jq
3. Send `notification` event with data data

## Port Configuration

The port in hook commands must match the configured server port. The `setup` command reads the current port from `~/.config/agent-pulse/config.yaml` (default: 8080) and uses it in all hook commands.

## Idempotency

Running `setup` multiple times:
- Replaces existing agent-pulse hook entries (matched by hook key: SessionStart, Stop, Notification)
- Does not create duplicate entries
- Preserves any other hooks the user has configured under different keys
