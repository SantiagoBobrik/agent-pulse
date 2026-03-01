package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/SantiagoBobrik/agent-pulse/internal/client"
	"github.com/SantiagoBobrik/agent-pulse/internal/domain"
)

type Dispatcher struct {
	clients []client.Client
}

func NewDispatcher(clients []client.Client) *Dispatcher {
	return &Dispatcher{clients: clients}
}

func (d *Dispatcher) Dispatch(event domain.Event) {
	var wg sync.WaitGroup
	for _, c := range d.clients {
		if !c.Accepts(event) {
			slog.Info("client skipped", "name", c.Name, "reason", "event_filtered")
			continue
		}
		wg.Add(1)
		go func(c client.Client) {
			defer wg.Done()
			if err := d.send(c, event); err != nil {
				slog.Error("client unreachable", "name", c.Name, "error", err)
			}
		}(c)
	}
	wg.Wait()
}

func (d *Dispatcher) send(c client.Client, event domain.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	timeout := c.Timeout
	if timeout == 0 {
		timeout = 2 * time.Second
	}

	httpClient := &http.Client{Timeout: timeout}

	req, err := http.NewRequest(http.MethodPost, c.URL, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	for k, v := range c.ResolvedHeaders() {
		req.Header.Set(k, v)
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("send to %s: %w", c.URL, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("client %s returned status %d", c.Name, resp.StatusCode)
	}

	slog.Info("client ok", "name", c.Name, "status", resp.StatusCode)
	return nil
}
