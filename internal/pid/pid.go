package pid

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/SantiagoBobrik/agent-pulse/internal/config"
)

func path() (string, error) {
	dir, err := config.Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "server.pid"), nil
}

func Write(p int) error {
	f, err := path()
	if err != nil {
		return err
	}
	return os.WriteFile(f, []byte(strconv.Itoa(p)), 0644)
}

func Read() (int, error) {
	f, err := path()
	if err != nil {
		return 0, err
	}
	data, err := os.ReadFile(f)
	if err != nil {
		return 0, fmt.Errorf("pid file not found: %w", err)
	}
	return strconv.Atoi(strings.TrimSpace(string(data)))
}

func Remove() {
	if f, err := path(); err == nil {
		os.Remove(f)
	}
}
