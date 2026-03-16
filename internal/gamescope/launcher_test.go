package gamescope

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/display"
)

func TestBuildArgsRejectsShellInjection(t *testing.T) {
	launcher := NewLauncher()
	launcher.TargetCommand = "sleep 1; rm -rf /"
	_, err := launcher.BuildArgs(display.DisplayConfig{Width: 1920, Height: 1080, RefreshHz: 60}, "card1-VIRTUAL-1")
	if err == nil {
		t.Fatal("expected error for shell injection in target command, got nil")
	}
}

func TestBuildArgsAllowsSafeCommand(t *testing.T) {
	launcher := NewLauncher()
	launcher.TargetCommand = "sleep infinity"
	_, err := launcher.BuildArgs(display.DisplayConfig{Width: 1920, Height: 1080, RefreshHz: 60}, "card1-VIRTUAL-1")
	if err != nil {
		t.Fatalf("unexpected error for safe target command: %v", err)
	}
}

func TestBuildArgsHDRDisabled(t *testing.T) {
	launcher := NewLauncher()

	args, err := launcher.BuildArgs(display.DisplayConfig{
		Width:     2560,
		Height:    1600,
		RefreshHz: 120,
		HDR:       false,
	}, "card1-VIRTUAL-1")
	if err != nil {
		t.Fatalf("BuildArgs returned error: %v", err)
	}

	expected := []string{
		"-O", "card1-VIRTUAL-1",
		"-W", "2560",
		"-H", "1600",
		"-r", "120",
		"--generate-drm-mode", "cvt",
		"--", "sh", "-lc", "sleep infinity",
	}

	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected args:\n%v\nexpected:\n%v", args, expected)
	}
}

func TestBuildArgsHDREnabled(t *testing.T) {
	launcher := NewLauncher()

	args, err := launcher.BuildArgs(display.DisplayConfig{
		Width:     2560,
		Height:    1600,
		RefreshHz: 120,
		HDR:       true,
	}, "card1-VIRTUAL-1")
	if err != nil {
		t.Fatalf("BuildArgs returned error: %v", err)
	}

	expected := []string{
		"-O", "card1-VIRTUAL-1",
		"-W", "2560",
		"-H", "1600",
		"-r", "120",
		"--generate-drm-mode", "cvt",
		"--hdr-enabled",
		"--", "sh", "-lc", "sleep infinity",
	}

	if !reflect.DeepEqual(args, expected) {
		t.Fatalf("unexpected args:\n%v\nexpected:\n%v", args, expected)
	}
}

func TestOpenLogFilePermissions(t *testing.T) {
	launcher := NewLauncher()
	launcher.LogPath = filepath.Join(t.TempDir(), "gamescope.log")

	f, err := launcher.openLogFile()
	if err != nil {
		t.Fatalf("openLogFile returned error: %v", err)
	}
	_ = f.Close()

	info, err := os.Stat(launcher.LogPath)
	if err != nil {
		t.Fatalf("stat log file: %v", err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("unexpected log permissions: got %o want 600", got)
	}
}
