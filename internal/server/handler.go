package server

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/SantiagoBobrik/agent-pulse/internal/domain"
)

func handleEvent(dispatcher *Dispatcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var eventPayload domain.Event
		if err := json.NewDecoder(r.Body).Decode(&eventPayload); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if !eventPayload.Type.IsValid() {
			http.Error(w, "unknown event type", http.StatusBadRequest)
			return
		}

		var data map[string]any
		if len(eventPayload.Data) > 0 {
			json.Unmarshal(eventPayload.Data, &data)
		}
		slog.Info("event received", "type", eventPayload.Type, "data", data)
		dispatcher.Dispatch(eventPayload)
		w.WriteHeader(http.StatusAccepted)
	}
}
