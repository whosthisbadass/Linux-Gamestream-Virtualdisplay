package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	t.Setenv("SVD_BACKEND", "portal")
	t.Setenv("SVD_GAMESCOPE_STARTUP_TIMEOUT_SEC", "15")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Backend != "portal" || cfg.GamescopeStartupTimeoutSec != 15 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestLoadFromConfigFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cfg.json")
	if err := os.WriteFile(path, []byte(`{"backend":"vkms","monitor_interval_sec":9}`), 0o644); err != nil {
		t.Fatal(err)
	}
	t.Setenv("SVD_CONFIG_FILE", path)
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.MonitorIntervalSec != 9 {
		t.Fatalf("unexpected monitor interval: %d", cfg.MonitorIntervalSec)
	}
}
