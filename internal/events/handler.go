package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/SantiagoBobrik/agent-pulse/internal/config"
	"github.com/SantiagoBobrik/agent-pulse/internal/domain"
	"golang.org/x/term"
)

// HandleEvent reads a JSON payload from stdin and dispatches it to the gateway
// as the given event type.
func HandleEvent(provider domain.Provider, eventType domain.EventType) error {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("no piped input on stdin (hooks are called by agentic tool, not manually)")
	}

	var data json.RawMessage
	if err := json.NewDecoder(os.Stdin).Decode(&data); err != nil {
		return fmt.Errorf("invalid stdin JSON: %w", err)
	}

	return dispatch(provider, eventType, data)
}

func dispatch(provider domain.Provider, eventType domain.EventType, data json.RawMessage) error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	body, err := json.Marshal(domain.Event{
		Provider: provider,
		Type:     eventType,
		Data:     data,
	})
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	url := fmt.Sprintf("http://localhost:%d/event", cfg.Port)
	client := &http.Client{Timeout: 2 * time.Second}

	resp, err := client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to send event to server: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("server returned status %d", resp.StatusCode)
	}

	return nil
}
