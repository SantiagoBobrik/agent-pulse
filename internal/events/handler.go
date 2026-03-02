package events

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/SantiagoBobrik/agent-pulse/internal/config"
	"github.com/SantiagoBobrik/agent-pulse/internal/domain"
	"github.com/SantiagoBobrik/agent-pulse/internal/logger"
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

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	if err := ensureServer(cfg); err != nil {
		return err
	}

	return dispatch(cfg, provider, eventType, data)
}

func ensureServer(cfg *config.Config) error {
	healthURL := fmt.Sprintf("%s:%d/health", cfg.GatewayURL, cfg.Port)
	client := &http.Client{Timeout: time.Second}

	if resp, err := client.Get(healthURL); err == nil {
		resp.Body.Close()
		return nil
	}

	logger.Info("server unreachable, starting it automatically")
	startServer()
	time.Sleep(time.Second)

	resp, err := client.Get(healthURL)
	if err != nil {
		return fmt.Errorf("server did not start: %w", err)
	}
	resp.Body.Close()
	return nil
}

func dispatch(cfg *config.Config, provider domain.Provider, eventType domain.EventType, data json.RawMessage) error {
	body, err := json.Marshal(domain.Event{
		Provider: provider,
		Type:     eventType,
		Data:     data,
	})
	if err != nil {
		return fmt.Errorf("marshal event: %w", err)
	}

	url := fmt.Sprintf("%s:%d/event", cfg.GatewayURL, cfg.Port)
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

func startServer() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	cmd := exec.Command(exe, "server", "start")
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		logger.Error("failed to start server", "error", err)
	}
}
