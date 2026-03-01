package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/SantiagoBobrik/agent-pulse/internal/client"
	"gopkg.in/yaml.v3"
)

const (
	DefaultPort        = 8080
	DefaultBindAddress = "127.0.0.1"
)

type Config struct {
	Port        int             `yaml:"port"`
	BindAddress string          `yaml:"bind_address"`
	GatewayURL  string          `yaml:"gateway_url,omitempty"`
	Clients     []client.Client `yaml:"clients,omitempty"`
}

func configPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "agent-pulse", "config.yaml"), nil
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:        DefaultPort,
		BindAddress: DefaultBindAddress,
	}

	path, err := configPath()
	if err != nil {
		return applyDefaults(cfg)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return applyDefaults(cfg)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid config at %s: %w", path, err)
	}

	if cfg.Port < 1024 || cfg.Port > 65535 {
		return nil, fmt.Errorf("port must be between 1024 and 65535, got %d", cfg.Port)
	}

	if cfg.BindAddress == "" {
		cfg.BindAddress = DefaultBindAddress
	}

	return applyDefaults(cfg)
}

func applyDefaults(cfg *Config) (*Config, error) {
	if cfg.GatewayURL == "" {
		cfg.GatewayURL = fmt.Sprintf("http://localhost:%d", cfg.Port)
	} else {
		if _, err := url.ParseRequestURI(cfg.GatewayURL); err != nil {
			return nil, fmt.Errorf("invalid gateway_url %q: %w", cfg.GatewayURL, err)
		}
	}
	return cfg, nil
}

func Save(cfg *Config) error {
	path, err := configPath()
	if err != nil {
		return fmt.Errorf("cannot determine config path: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("cannot create config directory: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("cannot marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("cannot write config: %w", err)
	}

	return nil
}
