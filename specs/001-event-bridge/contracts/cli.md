# CLI Contract

**Date**: 2026-02-25
**Branch**: `001-event-bridge`

## Commands

### `agent-pulse`

Root command. Displays help/usage when run without subcommands.

```
Usage:
  agent-pulse [command]

Available Commands:
  setup       Configure Claude Code hooks for the current project
  serve       Start the event bridge server

Flags:
  -h, --help   help for agent-pulse

Use "agent-pulse [command] --help" for more information about a command.
```

### `agent-pulse setup`

Configures the current project's `.claude/settings.json` with hooks that send lifecycle events to the agent-pulse server.

**Behavior**:
1. Reads `.claude/settings.json` (creates file and directory if missing)
2. Merges agent-pulse hooks into the `hooks` key
3. Preserves all existing settings and unrelated hooks
4. Writes updated settings back

**Output**:
- Success: `agent-pulse hooks configured in .claude/settings.json`
- Error (invalid JSON): `error: .claude/settings.json contains invalid JSON`

**Exit codes**:
- `0` — success
- `1` — error (file read/write failure, invalid JSON)

### `agent-pulse serve`

Starts the HTTP/WebSocket server.

**Flags**:
- `--port, -p` (int, default: from config or 8080) — server listen port

**Behavior**:
1. Loads config from `~/.config/agent-pulse/config.yaml`
2. Starts HTTP server on configured port
3. Listens for events on `POST /event`
4. Broadcasts events to connected WebSocket clients
5. Handles SIGINT/SIGTERM for graceful shutdown

**Output**:
```
[agent-pulse] server started on :8080
[agent-pulse] event received              type=session_start
[agent-pulse] broadcast to 2 client(s)    type=session_start
[agent-pulse] server shutting down...
```

**Exit codes**:
- `0` — clean shutdown
- `1` — error (port in use, config error)

**Error: port in use**:
```
error: port 8080 is already in use. Change the port in ~/.config/agent-pulse/config.yaml
```
