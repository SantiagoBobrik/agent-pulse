package server

import (
	"encoding/json"
	"net/http"

	"github.com/SantiagoBobrik/agent-pulse/internal/domain"
	"github.com/SantiagoBobrik/agent-pulse/internal/logger"
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

		logger.Info("event received", "type", eventPayload.Type, "provider", eventPayload.Provider)
		dispatcher.Dispatch(eventPayload)
		w.WriteHeader(http.StatusAccepted)
	}
}
