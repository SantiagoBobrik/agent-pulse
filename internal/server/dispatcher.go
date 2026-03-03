package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/SantiagoBobrik/agent-pulse/internal/client"
	"github.com/SantiagoBobrik/agent-pulse/internal/config"
	"github.com/SantiagoBobrik/agent-pulse/internal/domain"
	"github.com/SantiagoBobrik/agent-pulse/internal/logger"
)

type Dispatcher struct {
	broker *Broker
}

func NewDispatcher(broker *Broker) *Dispatcher {
	return &Dispatcher{broker: broker}
}

func (d *Dispatcher) Dispatch(event domain.Event) {
	cfg, err := config.Load()
	if err != nil {
		logger.Error("config reload failed", "error", err)
		return
	}

	var wg sync.WaitGroup
	for _, c := range cfg.Clients {
		if !c.Accepts(event) {
			logger.Info("client skipped", "name", c.Name, "reason", "event_filtered")
			continue
		}
		wg.Add(1)
		go func(c client.Client) {
			defer wg.Done()
			if err := d.send(c, event); err != nil {
				logger.Error("client unreachable", "name", c.Name, "error", err)
			}
		}(c)
	}
	wg.Wait()

	// Publish to SSE subscribers
	if d.broker != nil {
		d.broker.Publish(event)
	}
}

func (d *Dispatcher) send(c client.Client, event domain.Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	timeout := c.Timeout
	if timeout == 0 {
		timeout = 2000
	}

	httpClient := &http.Client{Timeout: time.Duration(timeout) * time.Millisecond}

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

	logger.Info("client ok", "name", c.Name, "status", resp.StatusCode)
	return nil
}
