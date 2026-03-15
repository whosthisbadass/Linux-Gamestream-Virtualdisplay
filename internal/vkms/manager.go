package vkms

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Manager struct {
	BasePath string
	DryRun   bool
	LogOut   io.Writer
}

type Instance struct {
	Name string
	Path string
}

func NewManager() *Manager {
	return &Manager{
		BasePath: "/sys/kernel/config/vkms",
		DryRun:   isTruthy(os.Getenv("SVD_DRY_RUN")),
		LogOut:   os.Stdout,
	}
}

func (m *Manager) Create(ctx context.Context, name string) (Instance, error) {
	if err := m.modprobeVKMS(ctx); err != nil {
		return Instance{}, err
	}

	instancePath := filepath.Join(m.BasePath, name)
	for _, dir := range []string{
		instancePath,
		filepath.Join(instancePath, "planes", "plane0"),
		filepath.Join(instancePath, "crtcs", "crtc0"),
		filepath.Join(instancePath, "encoders", "encoder0"),
		filepath.Join(instancePath, "connectors", "connector0"),
	} {
		if err := m.mkdir(dir); err != nil {
			return Instance{}, err
		}
	}

	if err := m.symlink(filepath.Join(instancePath, "crtcs", "crtc0"), filepath.Join(instancePath, "planes", "plane0", "possible_crtcs")); err != nil {
		return Instance{}, err
	}
	if err := m.symlink(filepath.Join(instancePath, "crtcs", "crtc0"), filepath.Join(instancePath, "encoders", "encoder0", "possible_crtcs")); err != nil {
		return Instance{}, err
	}
	if err := m.symlink(filepath.Join(instancePath, "encoders", "encoder0"), filepath.Join(instancePath, "connectors", "connector0", "possible_encoders")); err != nil {
		return Instance{}, err
	}

	if err := m.writeFile(filepath.Join(instancePath, "planes", "plane0", "type"), "1\n"); err != nil {
		return Instance{}, err
	}
	if err := m.writeFile(filepath.Join(instancePath, "connectors", "connector0", "status"), "1\n"); err != nil {
		return Instance{}, err
	}
	if err := m.writeFile(filepath.Join(instancePath, "enabled"), "1\n"); err != nil {
		return Instance{}, err
	}

	return Instance{Name: name, Path: instancePath}, nil
}

func (m *Manager) Destroy(name string) error {
	instancePath := filepath.Join(m.BasePath, name)
	if !m.DryRun {
		if _, err := os.Stat(instancePath); os.IsNotExist(err) {
			return nil
		} else if err != nil {
			return fmt.Errorf("stat vkms instance %s: %w", name, err)
		}
	}

	if err := m.writeFile(filepath.Join(instancePath, "enabled"), "0\n"); err != nil {
		return err
	}

	for _, link := range []string{
		filepath.Join(instancePath, "connectors", "connector0", "possible_encoders"),
		filepath.Join(instancePath, "encoders", "encoder0", "possible_crtcs"),
		filepath.Join(instancePath, "planes", "plane0", "possible_crtcs"),
	} {
		if err := m.remove(link); err != nil {
			return err
		}
	}

	for _, dir := range []string{
		filepath.Join(instancePath, "connectors", "connector0"),
		filepath.Join(instancePath, "encoders", "encoder0"),
		filepath.Join(instancePath, "crtcs", "crtc0"),
		filepath.Join(instancePath, "planes", "plane0"),
		filepath.Join(instancePath, "connectors"),
		filepath.Join(instancePath, "encoders"),
		filepath.Join(instancePath, "crtcs"),
		filepath.Join(instancePath, "planes"),
		instancePath,
	} {
		if err := m.rmdir(dir); err != nil {
			return err
		}
	}

	return nil
}

func (m *Manager) modprobeVKMS(ctx context.Context) error {
	if m.DryRun {
		m.logf("run modprobe vkms")
		return nil
	}
	if err := exec.CommandContext(ctx, "modprobe", "vkms").Run(); err != nil {
		return fmt.Errorf("modprobe vkms failed: %w", err)
	}
	return nil
}

func (m *Manager) mkdir(path string) error {
	if m.DryRun {
		m.logf("mkdir %s", path)
		return nil
	}
	if err := os.Mkdir(path, 0o755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("mkdir %s: %w", path, err)
	}
	return nil
}

func (m *Manager) symlink(target, link string) error {
	if m.DryRun {
		m.logf("ln -s %s %s", target, link)
		return nil
	}
	if err := os.Symlink(target, link); err != nil && !os.IsExist(err) {
		return fmt.Errorf("create symlink %s -> %s: %w", link, target, err)
	}
	return nil
}

func (m *Manager) writeFile(path, value string) error {
	if m.DryRun {
		m.logf("write %q to %s", strings.TrimSuffix(value, "\n"), path)
		return nil
	}
	if err := os.WriteFile(path, []byte(value), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func (m *Manager) remove(path string) error {
	if m.DryRun {
		m.logf("rm %s", path)
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %s: %w", path, err)
	}
	return nil
}

func (m *Manager) rmdir(path string) error {
	if m.DryRun {
		m.logf("rmdir %s", path)
		return nil
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("rmdir %s: %w", path, err)
	}
	return nil
}

func (m *Manager) logf(format string, args ...any) {
	if m.LogOut == nil {
		return
	}
	_, _ = fmt.Fprintf(m.LogOut, "DRY-RUN: "+format+"\n", args...)
}

func isTruthy(raw string) bool {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "1", "true", "yes", "on":
		return true
	default:
		return false
	}
}
