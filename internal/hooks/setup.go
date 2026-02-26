package hooks

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

func Setup(dir string, port int) error {
	claudeDir := filepath.Join(dir, ".claude")
	settingsPath := filepath.Join(claudeDir, "settings.json")

	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		return fmt.Errorf("cannot create .claude directory: %w", err)
	}

	settings := make(map[string]any)

	data, err := os.ReadFile(settingsPath)
	if err == nil {
		if err := json.Unmarshal(data, &settings); err != nil {
			return fmt.Errorf(".claude/settings.json contains invalid JSON")
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("cannot read .claude/settings.json: %w", err)
	}

	hooks := buildHooks(port)

	existingHooks, _ := settings["hooks"].(map[string]any)
	if existingHooks == nil {
		existingHooks = make(map[string]any)
	}
	for k, v := range hooks {
		existingHooks[k] = v
	}
	settings["hooks"] = existingHooks

	out, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal settings: %w", err)
	}

	if err := os.WriteFile(settingsPath, append(out, '\n'), 0644); err != nil {
		return fmt.Errorf("cannot write .claude/settings.json: %w", err)
	}

	return nil
}

func buildHooks(port int) map[string]any {
	base := fmt.Sprintf("http://localhost:%d", port)

	sessionStartCmd := fmt.Sprintf(
		`curl -sf %s/health > /dev/null 2>&1 || (claude-pulse serve > /dev/null 2>&1 &); sleep 1; curl -sf -X POST %s/event -H 'Content-Type: application/json' -d '{"type":"session_start"}'`,
		base, base,
	)

	stopCmd := fmt.Sprintf(
		`curl -sf -X POST %s/event -H 'Content-Type: application/json' -d '{"type":"stop"}'`,
		base,
	)

	notificationCmd := fmt.Sprintf(
		`jq -c '{type:"notification",extra:{message:.message,notification_type:.notification_type}}' | curl -sf -X POST %s/event -H 'Content-Type: application/json' -d @-`,
		base,
	)

	hookEntry := func(command string) []any {
		return []any{
			map[string]any{
				"hooks": []any{
					map[string]any{
						"type":    "command",
						"command": command,
					},
				},
			},
		}
	}

	return map[string]any{
		"SessionStart": hookEntry(sessionStartCmd),
		"Stop":         hookEntry(stopCmd),
		"Notification": hookEntry(notificationCmd),
	}
}
