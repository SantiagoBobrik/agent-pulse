package hooks

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
)

func Setup(dir string) error {
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

	hooks := buildHooks()

	existingHooks, _ := settings["hooks"].(map[string]any)
	if existingHooks == nil {
		existingHooks = make(map[string]any)
	}
	maps.Copy(existingHooks, hooks)
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

func buildHooks() map[string]any {
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
		"SessionStart": hookEntry("agent-pulse hook --provider claude --event session_start"),
		"SessionEnd":   hookEntry("agent-pulse hook --provider claude --event session_end"),
		"Stop":         hookEntry("agent-pulse hook --provider claude --event stop"),
		"Notification": hookEntry("agent-pulse hook --provider claude --event notification"),
	}
}
