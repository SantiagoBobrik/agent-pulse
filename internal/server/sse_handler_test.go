package server

import (
	"bufio"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/SantiagoBobrik/agent-pulse/internal/domain"
)

func TestHandleSSEStreamsEvents(t *testing.T) {
	broker := NewBroker()
	handler := handleSSE(broker)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	// Publish after a short delay so the handler has time to subscribe
	go func() {
		time.Sleep(100 * time.Millisecond)
		broker.Publish(domain.Event{Provider: domain.Providers.Claude, Type: domain.Events.Stop})
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		jsonPart, ok := strings.CutPrefix(line, "data: ")
		if !ok {
			continue
		}
		var got domain.Event
		if err := json.Unmarshal([]byte(jsonPart), &got); err != nil {
			t.Fatalf("invalid JSON: %v", err)
		}
		if got.Type != domain.Events.Stop {
			t.Errorf("event type = %q, want %q", got.Type, domain.Events.Stop)
		}
		return // success
	}
	t.Fatal("did not receive SSE data line")
}

func TestHandleSSEClientDisconnect(t *testing.T) {
	broker := NewBroker()
	handler := handleSSE(broker)

	srv := httptest.NewServer(handler)
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}

	// Cancel the context and close body to simulate client disconnect
	cancel()
	resp.Body.Close()

	// Give the handler time to detect the disconnect and clean up
	time.Sleep(100 * time.Millisecond)

	broker.mu.Lock()
	count := len(broker.subs)
	broker.mu.Unlock()

	if count != 0 {
		t.Errorf("expected 0 subscribers after disconnect, got %d", count)
	}
}
