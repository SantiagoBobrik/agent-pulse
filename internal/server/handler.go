package server

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

const (
	EventSessionStart = "session_start"
	EventStop         = "stop"
	EventNotification = "notification"
)

type Event struct {
	Type  string          `json:"type"`
	Extra json.RawMessage `json:"extra,omitempty"`
}

func isValidEventType(t string) bool {
	switch t {
	case EventSessionStart, EventStop, EventNotification:
		return true
	}
	return false
}

func handleEvent(hub *Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var event Event
		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if !isValidEventType(event.Type) {
			http.Error(w, "unknown event type", http.StatusBadRequest)
			return
		}

		data, err := json.Marshal(event)
		if err != nil {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}

		slog.Info("event received", "type", event.Type)
		hub.broadcast <- data
		w.WriteHeader(http.StatusAccepted)
	}
}
