package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/SantiagoBobrik/agent-pulse/internal/domain"
	"github.com/SantiagoBobrik/agent-pulse/internal/logger"
)

type Broker struct {
	mu   sync.Mutex
	subs map[http.ResponseWriter]*http.ResponseController
}

func NewBroker() *Broker {
	return &Broker{
		subs: make(map[http.ResponseWriter]*http.ResponseController),
	}
}

func (b *Broker) Subscribe(w http.ResponseWriter) (*http.ResponseController, func()) {
	rc := http.NewResponseController(w)

	b.mu.Lock()
	b.subs[w] = rc
	b.mu.Unlock()

	cleanup := func() {
		b.mu.Lock()
		delete(b.subs, w)
		b.mu.Unlock()
	}

	return rc, cleanup
}

func (b *Broker) Publish(event domain.Event) {
	data, err := json.Marshal(event)
	if err != nil {
		logger.Error("broker marshal failed", "error", err)
		return
	}

	msg := fmt.Appendf(nil, "data: %s\n\n", data)

	b.mu.Lock()
	defer b.mu.Unlock()

	for w, rc := range b.subs {
		if _, err := w.Write(msg); err == nil {
			rc.Flush()
		}
	}
}
