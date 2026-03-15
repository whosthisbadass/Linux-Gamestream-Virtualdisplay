package vkms

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestCreateDestroyPrivileged(t *testing.T) {
	if os.Getenv("SVD_PRIVILEGED_TESTS") != "1" {
		t.Skip("set SVD_PRIVILEGED_TESTS=1 to run privileged VKMS integration tests")
	}

	m := NewManager()
	m.DryRun = false
	name := "sunshine-it-" + time.Now().Format("150405")

	instance, err := m.Create(context.Background(), name)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if instance.Name != name {
		t.Fatalf("unexpected instance name %q", instance.Name)
	}

	if err := m.Destroy(name); err != nil {
		t.Fatalf("Destroy failed: %v", err)
	}
}
