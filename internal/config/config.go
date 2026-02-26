package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	DefaultPort = 8080
)

type Config struct {
	Port int `yaml:"port"`
}

func Load() (*Config, error) {
	cfg := &Config{Port: DefaultPort}

	home, err := os.UserHomeDir()
	if err != nil {
		return cfg, nil
	}

	path := filepath.Join(home, ".config", "claude-pulse", "config.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, nil
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("invalid config at %s: %w", path, err)
	}

	if cfg.Port < 1024 || cfg.Port > 65535 {
		return nil, fmt.Errorf("port must be between 1024 and 65535, got %d", cfg.Port)
	}

	return cfg, nil
}
