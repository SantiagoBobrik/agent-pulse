package projectconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFrom_FileFound(t *testing.T) {
	dir := t.TempDir()

	content := `{"metadata": {"project": "my-api", "team": "backend"}}`
	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(dir)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if string(cfg.Metadata) != `{"project": "my-api", "team": "backend"}` {
		t.Errorf("Metadata = %s, want %s", cfg.Metadata, `{"project": "my-api", "team": "backend"}`)
	}
}

func TestLoadFrom_FileMissing(t *testing.T) {
	dir := t.TempDir()

	cfg, err := LoadFrom(dir)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil config, got %+v", cfg)
	}
}

func TestLoadFrom_InvalidJSON(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, FileName), []byte("{{{invalid"), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(dir)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if cfg != nil {
		t.Errorf("expected nil config on error, got %+v", cfg)
	}
}

func TestLoadFrom_EmptyMetadata(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, FileName), []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadFrom(dir)
	if err != nil {
		t.Fatalf("LoadFrom() error = %v", err)
	}
	if cfg == nil {
		t.Fatal("expected non-nil config")
	}
	if cfg.Metadata != nil {
		t.Errorf("Metadata = %s, want nil", cfg.Metadata)
	}
}
