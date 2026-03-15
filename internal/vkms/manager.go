package vkms

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/clientdetector"
)

type Manager struct {
	BasePath string
}

type Instance struct {
	Name string
	Path string
}

func NewManager() *Manager {
	return &Manager{BasePath: "/sys/kernel/config/vkms"}
}

func (m *Manager) Create(ctx context.Context, name string, req clientdetector.ClientDisplayRequest) (Instance, error) {
	if err := exec.CommandContext(ctx, "modprobe", "vkms").Run(); err != nil {
		return Instance{}, fmt.Errorf("modprobe vkms failed: %w", err)
	}

	instancePath := filepath.Join(m.BasePath, name)
	if err := os.MkdirAll(instancePath, 0o755); err != nil {
		return Instance{}, fmt.Errorf("create vkms instance %s: %w", name, err)
	}

	if err := m.createPipeline(instancePath, req); err != nil {
		_ = os.RemoveAll(instancePath)
		return Instance{}, err
	}

	return Instance{Name: name, Path: instancePath}, nil
}

func (m *Manager) createPipeline(instancePath string, req clientdetector.ClientDisplayRequest) error {
	planePath := filepath.Join(instancePath, "planes", "plane-1")
	crtcPath := filepath.Join(instancePath, "crtcs", "crtc-1")
	encoderPath := filepath.Join(instancePath, "encoders", "encoder-1")
	connectorPath := filepath.Join(instancePath, "connectors", "connector-1")

	for _, path := range []string{planePath, crtcPath, encoderPath, connectorPath} {
		if err := os.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("create %s: %w", path, err)
		}
	}

	if err := ensureSymlink(filepath.Join(planePath, "crtc"), crtcPath); err != nil {
		return err
	}
	if err := ensureSymlink(filepath.Join(encoderPath, "crtc"), crtcPath); err != nil {
		return err
	}
	if err := ensureSymlink(filepath.Join(connectorPath, "encoder"), encoderPath); err != nil {
		return err
	}

	mode := fmt.Sprintf("%dx%d@%d", req.Width, req.Height, req.RefreshRate)
	modePath := filepath.Join(connectorPath, "mode")
	if err := writeIfExists(modePath, []byte(mode+"\n")); err != nil {
		return fmt.Errorf("write mode %s: %w", mode, err)
	}

	enabledPath := filepath.Join(connectorPath, "enabled")
	if err := writeIfExists(enabledPath, []byte("1\n")); err != nil {
		return fmt.Errorf("enable connector: %w", err)
	}

	return nil
}

func (m *Manager) Destroy(name string) error {
	instancePath := filepath.Join(m.BasePath, name)
	if _, err := os.Stat(instancePath); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(instancePath); err != nil {
		return fmt.Errorf("destroy vkms instance %s: %w", name, err)
	}
	return nil
}

func ensureSymlink(link, target string) error {
	if _, err := os.Lstat(link); err == nil {
		return nil
	}
	if err := os.Symlink(target, link); err != nil {
		return fmt.Errorf("create symlink %s -> %s: %w", link, target, err)
	}
	return nil
}

func writeIfExists(path string, data []byte) error {
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
