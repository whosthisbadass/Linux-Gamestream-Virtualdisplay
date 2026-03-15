package gamescope

import (
	"reflect"
	"testing"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/clientdetector"
)

func TestBuildArgsHDRDisabled(t *testing.T) {
	t.Setenv("SUNSHINE_VD_GAMESCOPE_TARGET", "sleep infinity")
	launcher := NewLauncher()

	args, err := launcher.BuildArgs(clientdetector.ClientDisplayRequest{
		Width:       2560,
		Height:      1600,
		RefreshRate: 120,
		HDR:         false,
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
	t.Setenv("SUNSHINE_VD_GAMESCOPE_TARGET", "sleep infinity")
	t.Setenv("SVD_GAMESCOPE_GENERATE_DRM_MODE", "cvt")
	launcher := NewLauncher()

	args, err := launcher.BuildArgs(clientdetector.ClientDisplayRequest{
		Width:       2560,
		Height:      1600,
		RefreshRate: 120,
		HDR:         true,
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
