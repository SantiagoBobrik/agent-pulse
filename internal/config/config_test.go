package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/SantiagoBobrik/agent-pulse/internal/client"
)

func TestLoadDefaults(t *testing.T) {
	// Point to a non-existent config path
	t.Setenv("HOME", t.TempDir())

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Port != DefaultPort {
		t.Errorf("Port = %d, want %d", cfg.Port, DefaultPort)
	}
	if cfg.BindAddress != DefaultBindAddress {
		t.Errorf("BindAddress = %q, want %q", cfg.BindAddress, DefaultBindAddress)
	}
	if len(cfg.Clients) != 0 {
		t.Errorf("Clients = %v, want empty", cfg.Clients)
	}
	if cfg.GatewayURL != DefaultGatewayURL {
		t.Errorf("GatewayURL = %q, want %q", cfg.GatewayURL, DefaultGatewayURL)
	}
}

func TestLoadWithClients(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	configDir := filepath.Join(dir, ".config", "agent-pulse")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		t.Fatal(err)
	}

	yaml := `port: 9090
bind_address: "0.0.0.0"
clients:
  - name: test-client
    url: http://192.168.1.100
    timeout: 3000
    events:
      - stop
      - notification
  - name: webhook
    url: https://hooks.example.com
    timeout: 5000
`
	if err := os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(yaml), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Port != 9090 {
		t.Errorf("Port = %d, want 9090", cfg.Port)
	}
	if cfg.BindAddress != "0.0.0.0" {
		t.Errorf("BindAddress = %q, want 0.0.0.0", cfg.BindAddress)
	}
	if len(cfg.Clients) != 2 {
		t.Fatalf("Clients count = %d, want 2", len(cfg.Clients))
	}
	if cfg.Clients[0].Name != "test-client" {
		t.Errorf("Client[0].Name = %q, want test-client", cfg.Clients[0].Name)
	}
	if cfg.Clients[0].Timeout != 3000 {
		t.Errorf("Client[0].Timeout = %v, want 3000", cfg.Clients[0].Timeout)
	}
	if len(cfg.Clients[0].Events) != 2 {
		t.Errorf("Client[0].Events count = %d, want 2", len(cfg.Clients[0].Events))
	}
}

func TestSaveAndRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	cfg := &Config{
		Port:        8080,
		BindAddress: "127.0.0.1",
		Clients: []client.Client{
			{
				Name:    "test",
				URL:     "http://example.com",
				Timeout: 2000,
				Events:  []string{"stop"},
			},
		},
	}

	if err := Save(cfg); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	loaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.Port != cfg.Port {
		t.Errorf("Port = %d, want %d", loaded.Port, cfg.Port)
	}
	if len(loaded.Clients) != 1 {
		t.Fatalf("Clients count = %d, want 1", len(loaded.Clients))
	}
	if loaded.Clients[0].Name != "test" {
		t.Errorf("Client.Name = %q, want test", loaded.Clients[0].Name)
	}
	if loaded.Clients[0].Timeout != 2000 {
		t.Errorf("Client.Timeout = %v, want 2000", loaded.Clients[0].Timeout)
	}
}

func TestLoadWithGatewayURL(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	configDir := filepath.Join(dir, ".config", "agent-pulse")
	os.MkdirAll(configDir, 0755)

	yaml := `port: 9090
gateway_url: "http://192.168.1.50"
`
	os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(yaml), 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.GatewayURL != "http://192.168.1.50" {
		t.Errorf("GatewayURL = %q, want http://192.168.1.50", cfg.GatewayURL)
	}
}

func TestLoadGatewayURLDefaultWhenOmitted(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	configDir := filepath.Join(dir, ".config", "agent-pulse")
	os.MkdirAll(configDir, 0755)

	yaml := `port: 3000
`
	os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(yaml), 0644)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.GatewayURL != DefaultGatewayURL {
		t.Errorf("GatewayURL = %q, want %q", cfg.GatewayURL, DefaultGatewayURL)
	}
}

func TestLoadInvalidGatewayURL(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	configDir := filepath.Join(dir, ".config", "agent-pulse")
	os.MkdirAll(configDir, 0755)

	yaml := `gateway_url: "not a url"
`
	os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte(yaml), 0644)

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid gateway_url")
	}
}

func TestLoadInvalidYAML(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	configDir := filepath.Join(dir, ".config", "agent-pulse")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("{{{invalid"), 0644)

	_, err := Load()
	if err == nil {
		t.Error("expected error for invalid YAML")
	}
}

func TestLoadInvalidPort(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("HOME", dir)

	configDir := filepath.Join(dir, ".config", "agent-pulse")
	os.MkdirAll(configDir, 0755)
	os.WriteFile(filepath.Join(configDir, "config.yaml"), []byte("port: 80"), 0644)

	_, err := Load()
	if err == nil {
		t.Error("expected error for port below 1024")
	}
}
