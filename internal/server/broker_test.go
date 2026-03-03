package server

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/SantiagoBobrik/agent-pulse/internal/domain"
)

func TestBrokerSubscribeAndPublish(t *testing.T) {
	b := NewBroker()
	rec := httptest.NewRecorder()
	_, cleanup := b.Subscribe(rec)
	defer cleanup()

	event := domain.Event{Provider: domain.Providers.Claude, Type: domain.Events.Stop}
	b.Publish(event)

	body := rec.Body.String()
	if !strings.HasPrefix(body, "data: ") {
		t.Errorf("expected SSE data prefix, got %q", body)
	}
	if !strings.HasSuffix(body, "\n\n") {
		t.Errorf("expected trailing double newline, got %q", body)
	}
	if !strings.Contains(body, `"event":"stop"`) {
		t.Errorf("expected stop event in body, got %q", body)
	}
}

func TestBrokerCleanup(t *testing.T) {
	b := NewBroker()
	rec := httptest.NewRecorder()
	_, cleanup := b.Subscribe(rec)
	cleanup()

	b.mu.Lock()
	count := len(b.subs)
	b.mu.Unlock()

	if count != 0 {
		t.Errorf("expected 0 subscribers after cleanup, got %d", count)
	}
}

func TestBrokerNoSubscribers(t *testing.T) {
	b := NewBroker()
	// Should not panic
	b.Publish(domain.Event{Type: domain.Events.Stop})
}

func TestBrokerMultipleSubscribers(t *testing.T) {
	b := NewBroker()

	const n = 5
	recorders := make([]*httptest.ResponseRecorder, n)
	for i := range n {
		recorders[i] = httptest.NewRecorder()
		_, cleanup := b.Subscribe(recorders[i])
		defer cleanup()
	}

	b.Publish(domain.Event{Type: domain.Events.Stop})

	for i, rec := range recorders {
		body := rec.Body.String()
		if !strings.HasPrefix(body, "data: ") {
			t.Errorf("subscriber %d: expected SSE data prefix, got %q", i, body)
		}
	}
}
