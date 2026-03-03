package server

import (
	"net/http"

	"github.com/SantiagoBobrik/agent-pulse/internal/logger"
)

func handleSSE(broker *Broker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		rc, cleanup := broker.Subscribe(w)
		defer cleanup()

		w.WriteHeader(http.StatusOK)
		rc.Flush()

		logger.Info("sse client connected", "remote", r.RemoteAddr)

		<-r.Context().Done()

		logger.Info("sse client disconnected", "remote", r.RemoteAddr)
	}
}
