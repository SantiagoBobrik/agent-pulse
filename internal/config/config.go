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
	DefaultPort        = 8789
	DefaultBindAddress = "127.0.0.1"
	DefaultGatewayURL  = "http://localhost"
)

type Config struct {
	Port        int             `yaml:"port"`
	BindAddress string          `yaml:"bind_address"`
	GatewayURL  string          `yaml:"gateway_url,omitempty"`
	Clients     []client.Client `yaml:"clients,omitempty"`
}

// PathOverride allows tests to redirect config.Load to a temp file.
var PathOverride string

// Dir returns the agent-pulse config directory: ~/.config/agent-pulse
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "agent-pulse"), nil
}

func configPath() (string, error) {
	if PathOverride != "" {
		return PathOverride, nil
	}
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yaml"), nil
}

func Load() (*Config, error) {
	cfg := &Config{}

	if path, err := configPath(); err == nil {
		if data, err := os.ReadFile(path); err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, fmt.Errorf("invalid config at %s: %w", path, err)
			}
		}
	}
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (c *Config) Validate() error {
	if c.Port == 0 {
		c.Port = DefaultPort
	}
	if c.Port < 1024 || c.Port > 65535 {
		return fmt.Errorf("port must be between 1024 and 65535, got %d", c.Port)
	}
	if c.BindAddress == "" {
		c.BindAddress = DefaultBindAddress
	}
	if c.GatewayURL == "" {
		c.GatewayURL = DefaultGatewayURL
	}
	if _, err := url.ParseRequestURI(c.GatewayURL); err != nil {
		return fmt.Errorf("invalid gateway_url %q: %w", c.GatewayURL, err)
	}
	return nil
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
