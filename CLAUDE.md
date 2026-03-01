# agent-pulse Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-02-25

## Active Technologies
- Go 1.25+ + cobra (CLI), chi/v5 (HTTP router), gorilla/websocket (WebSocket — existing), gopkg.in/yaml.v3 (config), slog (stdlib logging) (002-client-routing)
- YAML file at `~/.config/agent-pulse/config.yaml` (existing path) (002-client-routing)

- Go 1.25+ + cobra (CLI), chi (HTTP router), gorilla/websocket (WebSocket), slog (structured logging), gopkg.in/yaml.v3 (config) (001-event-bridge)

## Project Structure

```text
src/
tests/
```

## Commands

# Add commands for Go 1.24+

## Code Style

Go 1.24+: Follow standard conventions. Use `any` instead of `interface{}`. Use `slices.Contains` instead of manual for-range loops for membership checks.

## Recent Changes
- 002-client-routing: Added Go 1.24+ + cobra (CLI), chi/v5 (HTTP router), gorilla/websocket (WebSocket — existing), gopkg.in/yaml.v3 (config), slog (stdlib logging)

- 001-event-bridge: Added Go 1.24+ + cobra (CLI), chi (HTTP router), gorilla/websocket (WebSocket), slog (structured logging), gopkg.in/yaml.v3 (config)

<!-- MANUAL ADDITIONS START -->

## No Magic Strings

Never use raw string literals for values that have a named constant or struct field. Use the single source of truth defined in `internal/domain` for domain values (e.g., `domain.Events.Stop`). If a new domain value is needed, add it to the appropriate struct in `internal/domain` first, then reference it.

<!-- MANUAL ADDITIONS END -->
