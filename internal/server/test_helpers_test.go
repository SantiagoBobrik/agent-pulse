package server

import (
	"os"
	"testing"

	"github.com/SantiagoBobrik/agent-pulse/internal/client"
	"github.com/SantiagoBobrik/agent-pulse/internal/config"
	"gopkg.in/yaml.v3"
)

func newTestDispatcher(t *testing.T, clients []client.Client) *Dispatcher {
	t.Helper()

	cfg := &config.Config{Clients: clients}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.CreateTemp(t.TempDir(), "config-*.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.Write(data); err != nil {
		t.Fatal(err)
	}
	f.Close()

	config.PathOverride = f.Name()
	t.Cleanup(func() { config.PathOverride = "" })

	return NewDispatcher(nil)
}
