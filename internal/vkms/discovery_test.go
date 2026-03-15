package vkms

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverConnectorSingleConnected(t *testing.T) {
	root := t.TempDir()
	mustConnector(t, root, "card0-HDMI-A-1", "disconnected")
	mustConnector(t, root, "card1-VIRTUAL-1", "connected")

	connector, err := DiscoverConnector(root)
	if err != nil {
		t.Fatalf("DiscoverConnector returned error: %v", err)
	}
	if connector != "card1-VIRTUAL-1" {
		t.Fatalf("unexpected connector: %q", connector)
	}
}

func TestDiscoverConnectorDryRunForcedOverride(t *testing.T) {
	t.Setenv("SVD_DRY_RUN", "1")
	t.Setenv("SVD_FORCE_CONNECTOR", "card9-VIRTUAL-9")
	connector, err := DiscoverConnector(t.TempDir())
	if err != nil {
		t.Fatalf("DiscoverConnector returned error: %v", err)
	}
	if connector != "card9-VIRTUAL-9" {
		t.Fatalf("unexpected connector: %q", connector)
	}
}

func mustConnector(t *testing.T, root, name, status string) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "status"), []byte(status+"\n"), 0o644); err != nil {
		t.Fatalf("write status: %v", err)
	}
}
