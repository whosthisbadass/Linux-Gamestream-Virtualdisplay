package vkms

import (
	"bytes"
	"context"
	"path/filepath"
	"testing"
)

func TestManagerCreateDestroyDryRunOutput(t *testing.T) {
	var logs bytes.Buffer
	m := &Manager{BasePath: "/sys/kernel/config/vkms", DryRun: true, LogOut: &logs}

	instance, err := m.Create(context.Background(), "sunshine-test")
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if instance.Name != "sunshine-test" {
		t.Fatalf("unexpected instance name: %q", instance.Name)
	}

	if err := m.Destroy("sunshine-test"); err != nil {
		t.Fatalf("Destroy returned error: %v", err)
	}

	base := filepath.Join("/sys/kernel/config/vkms", "sunshine-test")
	expected := "" +
		"DRY-RUN: run modprobe vkms\n" +
		"DRY-RUN: mkdir " + base + "\n" +
		"DRY-RUN: mkdir " + filepath.Join(base, "planes", "plane0") + "\n" +
		"DRY-RUN: mkdir " + filepath.Join(base, "crtcs", "crtc0") + "\n" +
		"DRY-RUN: mkdir " + filepath.Join(base, "encoders", "encoder0") + "\n" +
		"DRY-RUN: mkdir " + filepath.Join(base, "connectors", "connector0") + "\n" +
		"DRY-RUN: ln -s " + filepath.Join(base, "crtcs", "crtc0") + " " + filepath.Join(base, "planes", "plane0", "possible_crtcs") + "\n" +
		"DRY-RUN: ln -s " + filepath.Join(base, "crtcs", "crtc0") + " " + filepath.Join(base, "encoders", "encoder0", "possible_crtcs") + "\n" +
		"DRY-RUN: ln -s " + filepath.Join(base, "encoders", "encoder0") + " " + filepath.Join(base, "connectors", "connector0", "possible_encoders") + "\n" +
		"DRY-RUN: write \"1\" to " + filepath.Join(base, "planes", "plane0", "type") + "\n" +
		"DRY-RUN: write \"1\" to " + filepath.Join(base, "connectors", "connector0", "status") + "\n" +
		"DRY-RUN: write \"1\" to " + filepath.Join(base, "enabled") + "\n" +
		"DRY-RUN: write \"0\" to " + filepath.Join(base, "enabled") + "\n" +
		"DRY-RUN: rm " + filepath.Join(base, "connectors", "connector0", "possible_encoders") + "\n" +
		"DRY-RUN: rm " + filepath.Join(base, "encoders", "encoder0", "possible_crtcs") + "\n" +
		"DRY-RUN: rm " + filepath.Join(base, "planes", "plane0", "possible_crtcs") + "\n" +
		"DRY-RUN: rmdir " + filepath.Join(base, "connectors", "connector0") + "\n" +
		"DRY-RUN: rmdir " + filepath.Join(base, "encoders", "encoder0") + "\n" +
		"DRY-RUN: rmdir " + filepath.Join(base, "crtcs", "crtc0") + "\n" +
		"DRY-RUN: rmdir " + filepath.Join(base, "planes", "plane0") + "\n" +
		"DRY-RUN: rmdir " + filepath.Join(base, "connectors") + "\n" +
		"DRY-RUN: rmdir " + filepath.Join(base, "encoders") + "\n" +
		"DRY-RUN: rmdir " + filepath.Join(base, "crtcs") + "\n" +
		"DRY-RUN: rmdir " + filepath.Join(base, "planes") + "\n" +
		"DRY-RUN: rmdir " + base + "\n"

	if logs.String() != expected {
		t.Fatalf("unexpected dry-run output:\n%s\nexpected:\n%s", logs.String(), expected)
	}
}
