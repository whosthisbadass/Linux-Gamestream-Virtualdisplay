package cleanup

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type SessionState struct {
	InstanceName string `json:"instance_name"`
	GamescopePID int    `json:"gamescope_pid"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	FPS          int    `json:"fps"`
	HDR          bool   `json:"hdr"`
}

func StateFilePath() string {
	base := os.Getenv("XDG_RUNTIME_DIR")
	if base == "" {
		base = "/tmp"
	}
	return filepath.Join(base, "sunshine-virtual-display", "session-state.json")
}

func Save(state SessionState) error {
	path := StateFilePath()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create state dir: %w", err)
	}

	payload, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal state: %w", err)
	}

	if err := os.WriteFile(path, payload, 0o644); err != nil {
		return fmt.Errorf("write state file: %w", err)
	}

	return nil
}

func Load() (SessionState, error) {
	path := StateFilePath()
	payload, err := os.ReadFile(path)
	if err != nil {
		return SessionState{}, fmt.Errorf("read state file: %w", err)
	}

	var state SessionState
	if err := json.Unmarshal(payload, &state); err != nil {
		return SessionState{}, fmt.Errorf("parse state file: %w", err)
	}

	return state, nil
}

func Remove() error {
	if err := os.Remove(StateFilePath()); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
