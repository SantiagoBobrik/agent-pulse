package hooks

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestSetupFreshDirectory(t *testing.T) {
	dir := t.TempDir()

	if err := Setup(dir); err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}

	var settings map[string]any
	if err := json.Unmarshal(data, &settings); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	hooks, ok := settings["hooks"].(map[string]any)
	if !ok {
		t.Fatal("hooks key missing or not a map")
	}

	expected := map[string]string{
		"SessionStart": "agent-pulse hook session_start",
		"Stop":         "agent-pulse hook stop",
		"Notification": "agent-pulse hook notification",
	}

	for hookName, wantCmd := range expected {
		entries, ok := hooks[hookName].([]any)
		if !ok || len(entries) == 0 {
			t.Errorf("hook %q missing or empty", hookName)
			continue
		}
		entry := entries[0].(map[string]any)
		hooksList := entry["hooks"].([]any)
		hook := hooksList[0].(map[string]any)
		if hook["command"] != wantCmd {
			t.Errorf("hook %q command = %q, want %q", hookName, hook["command"], wantCmd)
		}
		if hook["type"] != "command" {
			t.Errorf("hook %q type = %q, want %q", hookName, hook["type"], "command")
		}
	}
}

func TestSetupPreservesExistingHooks(t *testing.T) {
	dir := t.TempDir()
	claudeDir := filepath.Join(dir, ".claude")
	os.MkdirAll(claudeDir, 0755)

	existing := map[string]any{
		"hooks": map[string]any{
			"CustomHook": []any{
				map[string]any{
					"hooks": []any{
						map[string]any{"type": "command", "command": "echo custom"},
					},
				},
			},
		},
		"otherSetting": "preserved",
	}

	data, _ := json.MarshalIndent(existing, "", "  ")
	os.WriteFile(filepath.Join(claudeDir, "settings.json"), data, 0644)

	if err := Setup(dir); err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	out, _ := os.ReadFile(filepath.Join(claudeDir, "settings.json"))
	var settings map[string]any
	json.Unmarshal(out, &settings)

	if settings["otherSetting"] != "preserved" {
		t.Error("otherSetting was not preserved")
	}

	hooks := settings["hooks"].(map[string]any)
	if _, ok := hooks["CustomHook"]; !ok {
		t.Error("CustomHook was not preserved")
	}
	if _, ok := hooks["SessionStart"]; !ok {
		t.Error("SessionStart was not added")
	}
}

func TestSetupIdempotent(t *testing.T) {
	dir := t.TempDir()

	if err := Setup(dir); err != nil {
		t.Fatalf("first Setup() error: %v", err)
	}

	first, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))

	if err := Setup(dir); err != nil {
		t.Fatalf("second Setup() error: %v", err)
	}

	second, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))

	if string(first) != string(second) {
		t.Error("Setup is not idempotent: output differs between runs")
	}
}

func TestSetupUsesGoBinaryFormat(t *testing.T) {
	dir := t.TempDir()

	if err := Setup(dir); err != nil {
		t.Fatalf("Setup() error: %v", err)
	}

	data, _ := os.ReadFile(filepath.Join(dir, ".claude", "settings.json"))
	content := string(data)

	// Must use Go binary commands, not curl/jq
	if contains(content, "curl") {
		t.Error("hook commands should not contain curl")
	}
	if contains(content, "jq") {
		t.Error("hook commands should not contain jq")
	}
	if !contains(content, "agent-pulse hook") {
		t.Error("hook commands should use 'agent-pulse hook' format")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
