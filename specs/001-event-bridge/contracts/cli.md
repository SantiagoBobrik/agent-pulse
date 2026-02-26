# CLI Contract

**Date**: 2026-02-25
**Branch**: `001-event-bridge`

## Commands

### `claude-pulse`

Root command. Displays help/usage when run without subcommands.

```
Usage:
  claude-pulse [command]

Available Commands:
  setup       Configure Claude Code hooks for the current project
  serve       Start the event bridge server

Flags:
  -h, --help   help for claude-pulse

Use "claude-pulse [command] --help" for more information about a command.
```

### `claude-pulse setup`

Configures the current project's `.claude/settings.json` with hooks that send lifecycle events to the claude-pulse server.

**Behavior**:
1. Reads `.claude/settings.json` (creates file and directory if missing)
2. Merges claude-pulse hooks into the `hooks` key
3. Preserves all existing settings and unrelated hooks
4. Writes updated settings back

**Output**:
- Success: `claude-pulse hooks configured in .claude/settings.json`
- Error (invalid JSON): `error: .claude/settings.json contains invalid JSON`

**Exit codes**:
- `0` — success
- `1` — error (file read/write failure, invalid JSON)

### `claude-pulse serve`

Starts the HTTP/WebSocket server.

**Flags**:
- `--port, -p` (int, default: from config or 8080) — server listen port

**Behavior**:
1. Loads config from `~/.config/claude-pulse/config.yaml`
2. Starts HTTP server on configured port
3. Listens for events on `POST /event`
4. Broadcasts events to connected WebSocket clients
5. Handles SIGINT/SIGTERM for graceful shutdown

**Output**:
```
[claude-pulse] server started on :8080
[claude-pulse] event received              type=session_start
[claude-pulse] broadcast to 2 client(s)    type=session_start
[claude-pulse] server shutting down...
```

**Exit codes**:
- `0` — clean shutdown
- `1` — error (port in use, config error)

**Error: port in use**:
```
error: port 8080 is already in use. Change the port in ~/.config/claude-pulse/config.yaml
```
