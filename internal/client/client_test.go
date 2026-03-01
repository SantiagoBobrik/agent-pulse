package client

import (
	"testing"
	"time"

	"github.com/SantiagoBobrik/agent-pulse/internal/domain"
)

func TestClientValidate(t *testing.T) {
	tests := []struct {
		name    string
		client  Client
		wantErr bool
	}{
		{
			name: "valid client with all fields",
			client: Client{
				Name:    "escritorio",
				URL:     "http://192.168.1.100",
				Timeout: 2 * time.Second,
				Events:  []string{"stop", "notification"},
			},
			wantErr: false,
		},
		{
			name: "valid client with no scheme prepends http",
			client: Client{
				Name: "test",
				URL:  "192.168.1.100",
			},
			wantErr: false,
		},
		{
			name: "valid client with empty events receives all",
			client: Client{
				Name: "all-events",
				URL:  "http://example.com",
			},
			wantErr: false,
		},
		{
			name:    "empty name",
			client:  Client{URL: "http://example.com"},
			wantErr: true,
		},
		{
			name:    "invalid name characters",
			client:  Client{Name: "has spaces", URL: "http://example.com"},
			wantErr: true,
		},
		{
			name:    "empty URL",
			client:  Client{Name: "test"},
			wantErr: true,
		},
		{
			name: "timeout too large",
			client: Client{
				Name:    "test",
				URL:     "http://example.com",
				Timeout: 60 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "negative timeout",
			client: Client{
				Name:    "test",
				URL:     "http://example.com",
				Timeout: -1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid event type",
			client: Client{
				Name:   "test",
				URL:    "http://example.com",
				Events: []string{"invalid_event"},
			},
			wantErr: true,
		},
		{
			name: "valid with custom headers",
			client: Client{
				Name:    "authed",
				URL:     "http://example.com",
				Headers: map[string]string{"Authorization": "Bearer secret"},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.client.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClientAccepts(t *testing.T) {
	tests := []struct {
		name      string
		events    []string
		eventType domain.EventType
		want      bool
	}{
		{
			name:      "empty events accepts all",
			events:    nil,
			eventType: domain.Events.Stop,
			want:      true,
		},
		{
			name:      "subscribed event accepted",
			events:    []string{"stop", "notification"},
			eventType: domain.Events.Stop,
			want:      true,
		},
		{
			name:      "unsubscribed event rejected",
			events:    []string{"stop"},
			eventType: domain.Events.Notification,
			want:      false,
		},
		{
			name:      "empty slice accepts all",
			events:    []string{},
			eventType: domain.Events.SessionStart,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Events: tt.events}
			if got := c.Accepts(tt.eventType); got != tt.want {
				t.Errorf("Accepts(%q) = %v, want %v", tt.eventType, got, tt.want)
			}
		})
	}
}

func TestIsUniqueName(t *testing.T) {
	existing := []Client{
		{Name: "escritorio"},
		{Name: "slack"},
	}

	if IsUniqueName("escritorio", existing) {
		t.Error("expected escritorio to not be unique")
	}
	if IsUniqueName("Escritorio", existing) {
		t.Error("expected case-insensitive match")
	}
	if !IsUniqueName("new-client", existing) {
		t.Error("expected new-client to be unique")
	}
}
