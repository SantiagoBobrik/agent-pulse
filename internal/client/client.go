package client

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/SantiagoBobrik/agent-pulse/internal/domain"
)

var validEventTypes = map[domain.EventType]bool{
	domain.Events.SessionStart: true,
	domain.Events.SessionEnd:   true,
	domain.Events.Stop:         true,
	domain.Events.Notification: true,
}

var namePattern = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)

type Client struct {
	Name      string            `yaml:"name" json:"name"`
	URL       string            `yaml:"url" json:"url"`
	Timeout   time.Duration     `yaml:"timeout" json:"timeout"`
	Events    []string          `yaml:"events,omitempty" json:"events,omitempty"`
	Providers []string          `yaml:"providers,omitempty" json:"providers,omitempty"`
	Headers   map[string]string `yaml:"headers,omitempty" json:"headers,omitempty"`
}

func (c *Client) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("client name is required")
	}
	if !namePattern.MatchString(c.Name) {
		return fmt.Errorf("client name must be alphanumeric (hyphens and underscores allowed): %q", c.Name)
	}

	if c.URL == "" {
		return fmt.Errorf("client URL is required")
	}
	if !strings.Contains(c.URL, "://") {
		c.URL = "http://" + c.URL
	}
	if _, err := url.ParseRequestURI(c.URL); err != nil {
		return fmt.Errorf("invalid client URL %q: %w", c.URL, err)
	}

	if c.Timeout == 0 {
		c.Timeout = 2 * time.Second
	}
	if c.Timeout < 0 || c.Timeout > 30*time.Second {
		return fmt.Errorf("timeout must be between 0 and 30s, got %s", c.Timeout)
	}

	for _, e := range c.Events {
		if !validEventTypes[domain.EventType(e)] {
			return fmt.Errorf("invalid event type %q (valid: session_start, session_end, stop, notification)", e)
		}
	}

	return nil
}

func (c *Client) Accepts(event domain.Event) bool {
	acceptEvent := len(c.Events) == 0 || slices.Contains(c.Events,
		event.Type.String())
	acceptProvider := len(c.Providers) == 0 ||
		slices.Contains(c.Providers, event.Provider.String())
	return acceptEvent && acceptProvider
}

func (c *Client) ResolvedHeaders() map[string]string {
	resolved := make(map[string]string, len(c.Headers))
	for k, v := range c.Headers {
		resolved[k] = resolveEnvVar(v)
	}
	return resolved
}

func resolveEnvVar(val string) string {
	if strings.HasPrefix(val, "${") && strings.HasSuffix(val, "}") {
		envName := val[2 : len(val)-1]
		if envVal := os.Getenv(envName); envVal != "" {
			return envVal
		}
	}
	return val
}

func IsUniqueName(name string, existing []Client) bool {
	for _, c := range existing {
		if strings.EqualFold(c.Name, name) {
			return false
		}
	}
	return true
}
