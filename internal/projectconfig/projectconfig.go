package projectconfig

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const FileName = ".agent-pulse.json"

type ProjectConfig struct {
	Metadata json.RawMessage `json:"metadata"`
}

// Load reads .agent-pulse.json from the current working directory.
// Returns nil, nil if the file does not exist.
// Returns an error only for malformed JSON.
func Load() (*ProjectConfig, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("get working directory: %w", err)
	}

	return LoadFrom(wd)
}

// LoadFrom reads .agent-pulse.json from the given directory.
// Returns nil, nil if the file does not exist.
// Returns an error only for malformed JSON.
func LoadFrom(dir string) (*ProjectConfig, error) {
	data, err := os.ReadFile(filepath.Join(dir, FileName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, fmt.Errorf("read %s: %w", FileName, err)
	}

	var cfg ProjectConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse %s: %w", FileName, err)
	}

	return &cfg, nil
}
