package server

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/SantiagoBobrik/agent-pulse/internal/client"
	"github.com/SantiagoBobrik/agent-pulse/internal/domain"
)

func TestDispatchFanOut(t *testing.T) {
	var count atomic.Int32
	ts1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		w.WriteHeader(200)
	}))
	defer ts1.Close()

	ts2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count.Add(1)
		w.WriteHeader(200)
	}))
	defer ts2.Close()

	d := NewDispatcher([]client.Client{
		{Name: "c1", URL: ts1.URL, Timeout: 2000},
		{Name: "c2", URL: ts2.URL, Timeout: 2000},
	})

	d.Dispatch(domain.Event{Type: "stop", Data: json.RawMessage(`{"session_id":"test"}`)})

	if count.Load() != 2 {
		t.Errorf("expected 2 deliveries, got %d", count.Load())
	}
}

func TestDispatchEventFiltering(t *testing.T) {
	var received atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received.Add(1)
		w.WriteHeader(200)
	}))
	defer ts.Close()

	d := NewDispatcher([]client.Client{
		{Name: "stop-only", URL: ts.URL, Timeout: 2000, Events: []string{"stop"}},
	})

	d.Dispatch(domain.Event{Type: "session_start"})

	if received.Load() != 0 {
		t.Errorf("expected 0 deliveries for filtered event, got %d", received.Load())
	}

	d.Dispatch(domain.Event{Type: "stop"})

	if received.Load() != 1 {
		t.Errorf("expected 1 delivery for matching event, got %d", received.Load())
	}
}

func TestDispatchEmptyEventsAcceptsAll(t *testing.T) {
	var received atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received.Add(1)
		w.WriteHeader(200)
	}))
	defer ts.Close()

	d := NewDispatcher([]client.Client{
		{Name: "all", URL: ts.URL, Timeout: 2000},
	})

	d.Dispatch(domain.Event{Type: "session_start"})
	d.Dispatch(domain.Event{Type: "stop"})
	d.Dispatch(domain.Event{Type: "notification"})

	if received.Load() != 3 {
		t.Errorf("expected 3 deliveries, got %d", received.Load())
	}
}

func TestDispatchClientTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(3 * time.Second)
		w.WriteHeader(200)
	}))
	defer ts.Close()

	d := NewDispatcher([]client.Client{
		{Name: "slow", URL: ts.URL, Timeout: 100},
	})

	start := time.Now()
	d.Dispatch(domain.Event{Type: "stop"})
	elapsed := time.Since(start)

	if elapsed > 1*time.Second {
		t.Errorf("dispatch took %v, expected <1s (timeout should have fired)", elapsed)
	}
}

func TestDispatchClientUnreachable(t *testing.T) {
	// Use a port that's not listening
	d := NewDispatcher([]client.Client{
		{Name: "down", URL: "http://127.0.0.1:1", Timeout: 500},
	})

	// Should not panic
	d.Dispatch(domain.Event{Type: "stop"})
}

func TestDispatchClientNon2xx(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
	}))
	defer ts.Close()

	d := NewDispatcher([]client.Client{
		{Name: "error", URL: ts.URL, Timeout: 2000},
	})

	// Should not panic, error is logged
	d.Dispatch(domain.Event{Type: "stop"})
}

func TestDispatchCustomHeaders(t *testing.T) {
	var gotAuth, gotAPIKey string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotAPIKey = r.Header.Get("X-API-Key")
		w.WriteHeader(200)
	}))
	defer ts.Close()

	d := NewDispatcher([]client.Client{
		{
			Name:    "authed",
			URL:     ts.URL,
			Timeout: 2000,
			Headers: map[string]string{
				"Authorization": "Bearer my-secret",
				"X-API-Key":     "key-123",
			},
		},
	})

	d.Dispatch(domain.Event{Type: "stop"})

	if gotAuth != "Bearer my-secret" {
		t.Errorf("Authorization = %q, want %q", gotAuth, "Bearer my-secret")
	}
	if gotAPIKey != "key-123" {
		t.Errorf("X-API-Key = %q, want %q", gotAPIKey, "key-123")
	}
}

func TestDispatchOneFailureDoesNotBlockOthers(t *testing.T) {
	var okReceived atomic.Int32
	okServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		okReceived.Add(1)
		w.WriteHeader(200)
	}))
	defer okServer.Close()

	slowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(200)
	}))
	defer slowServer.Close()

	d := NewDispatcher([]client.Client{
		{Name: "slow", URL: slowServer.URL, Timeout: 100},
		{Name: "fast", URL: okServer.URL, Timeout: 2000},
	})

	start := time.Now()
	d.Dispatch(domain.Event{Type: "stop"})
	elapsed := time.Since(start)

	if okReceived.Load() != 1 {
		t.Error("fast client should have received the event")
	}
	if elapsed > 1*time.Second {
		t.Errorf("dispatch took %v, should have completed quickly", elapsed)
	}
}

func TestDispatchPayload(t *testing.T) {
	var receivedBody []byte
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		w.WriteHeader(200)
	}))
	defer ts.Close()

	d := NewDispatcher([]client.Client{
		{Name: "test", URL: ts.URL, Timeout: 2000},
	})

	d.Dispatch(domain.Event{Type: "stop", Data: json.RawMessage(`{"session_id":"abc","message":"done"}`)})

	var event map[string]any
	if err := json.Unmarshal(receivedBody, &event); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if event["event"] != "stop" {
		t.Errorf("event = %v, want stop", event["event"])
	}
	data, ok := event["data"].(map[string]any)
	if !ok {
		t.Fatal("data is not a map")
	}
	if data["session_id"] != "abc" {
		t.Errorf("session_id = %v, want abc", data["session_id"])
	}
}
