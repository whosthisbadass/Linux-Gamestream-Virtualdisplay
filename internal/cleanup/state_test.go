package cleanup

import (
	"os"
	"testing"
)

func TestLockAcquireRelease(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	lock, err := AcquireLock()
	if err != nil {
		t.Fatalf("acquire lock: %v", err)
	}
	if _, err := AcquireLock(); err == nil {
		t.Fatalf("expected second lock acquire to fail")
	}
	if err := ReleaseLock(lock); err != nil {
		t.Fatalf("release lock: %v", err)
	}
	if _, err := os.Stat(LockFilePath()); !os.IsNotExist(err) {
		t.Fatalf("lock path still exists")
	}
}

func TestSaveLoadState(t *testing.T) {
	t.Setenv("XDG_RUNTIME_DIR", t.TempDir())
	in := SessionState{InstanceName: "a", GamescopePID: 42, Connector: "card1-VIRTUAL-1", Width: 1920, Height: 1080, FPS: 60, HDR: false, Backend: "vkms"}
	if err := Save(in); err != nil {
		t.Fatalf("save: %v", err)
	}
	out, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if out.GamescopePID != in.GamescopePID || out.Backend != "vkms" {
		t.Fatalf("unexpected loaded state: %+v", out)
	}
}
