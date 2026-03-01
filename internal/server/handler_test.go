package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/SantiagoBobrik/agent-pulse/internal/client"
)

func TestHandleEventValid(t *testing.T) {
	var dispatched atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		dispatched.Add(1)
		w.WriteHeader(200)
	}))
	defer ts.Close()

	d := NewDispatcher([]client.Client{
		{Name: "test", URL: ts.URL, Timeout: 2000},
	})

	handler := handleEvent(d)

	body := `{"event":"stop","data":{"session_id":"abc"}}`
	req := httptest.NewRequest(http.MethodPost, "/event", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusAccepted)
	}
	if dispatched.Load() != 1 {
		t.Error("event should have been dispatched")
	}
}

func TestHandleEventInvalidJSON(t *testing.T) {
	d := NewDispatcher(nil)
	handler := handleEvent(d)

	req := httptest.NewRequest(http.MethodPost, "/event", bytes.NewBufferString("{invalid"))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleEventUnknownType(t *testing.T) {
	d := NewDispatcher(nil)
	handler := handleEvent(d)

	body := `{"event":"unknown"}`
	req := httptest.NewRequest(http.MethodPost, "/event", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	handler(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestHandleEventAllTypes(t *testing.T) {
	types := []string{"session_start", "stop", "notification"}
	for _, et := range types {
		t.Run(et, func(t *testing.T) {
			d := NewDispatcher(nil)
			handler := handleEvent(d)

			event := map[string]any{"event": et}
			data, _ := json.Marshal(event)

			req := httptest.NewRequest(http.MethodPost, "/event", bytes.NewReader(data))
			rec := httptest.NewRecorder()

			handler(rec, req)

			if rec.Code != http.StatusAccepted {
				t.Errorf("status = %d, want %d for type %q", rec.Code, http.StatusAccepted, et)
			}
		})
	}
}
