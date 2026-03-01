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
		providers []string
		event     domain.Event
		want      bool
	}{
		{
			name:  "empty events and providers accepts all",
			event: domain.Event{Type: domain.Events.Stop, Provider: domain.Providers.Claude},
			want:  true,
		},
		{
			name:   "subscribed event accepted",
			events: []string{"stop", "notification"},
			event:  domain.Event{Type: domain.Events.Stop, Provider: domain.Providers.Claude},
			want:   true,
		},
		{
			name:   "unsubscribed event rejected",
			events: []string{"stop"},
			event:  domain.Event{Type: domain.Events.Notification, Provider: domain.Providers.Claude},
			want:   false,
		},
		{
			name:  "empty events accepts all",
			event: domain.Event{Type: domain.Events.SessionStart, Provider: domain.Providers.Claude},
			want:  true,
		},
		{
			name:      "subscribed provider accepted",
			providers: []string{"claude"},
			event:     domain.Event{Type: domain.Events.Stop, Provider: domain.Providers.Claude},
			want:      true,
		},
		{
			name:      "unsubscribed provider rejected",
			providers: []string{"claude"},
			event:     domain.Event{Type: domain.Events.Stop, Provider: domain.Providers.Gemini},
			want:      false,
		},
		{
			name:  "empty providers accepts all providers",
			event: domain.Event{Type: domain.Events.Stop, Provider: domain.Providers.Gemini},
			want:  true,
		},
		{
			name:      "both filters match",
			events:    []string{"stop"},
			providers: []string{"claude"},
			event:     domain.Event{Type: domain.Events.Stop, Provider: domain.Providers.Claude},
			want:      true,
		},
		{
			name:      "event matches but provider does not",
			events:    []string{"stop"},
			providers: []string{"claude"},
			event:     domain.Event{Type: domain.Events.Stop, Provider: domain.Providers.Gemini},
			want:      false,
		},
		{
			name:      "provider matches but event does not",
			events:    []string{"stop"},
			providers: []string{"claude"},
			event:     domain.Event{Type: domain.Events.Notification, Provider: domain.Providers.Claude},
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{Events: tt.events, Providers: tt.providers}
			if got := c.Accepts(tt.event); got != tt.want {
				t.Errorf("Accepts(%v) = %v, want %v", tt.event, got, tt.want)
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
