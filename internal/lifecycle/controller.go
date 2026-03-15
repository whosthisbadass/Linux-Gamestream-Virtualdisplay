package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/backend"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/cleanup"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/clientdetector"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/config"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/display"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/gamescope"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/rules"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/vkms"
	"gopkg.in/yaml.v3"
)

type Controller struct {
	cfg       config.Config
	backend   backend.Backend
	Gamescope *gamescope.Launcher
}

func NewController() *Controller {
	cfg, _ := config.Load()
	vkmsManager := vkms.NewManager()
	vkmsManager.DryRun = cfg.DryRun
	b := pickBackend(cfg, vkmsManager)
	launcher := gamescope.NewLauncher()
	launcher.LogPath = cfg.GamescopeLogPath
	launcher.TargetCommand = cfg.GamescopeTarget
	launcher.ModeGeneration = cfg.GamescopeGenerateDRMMode
	launcher.StartupTimeout = cfg.GamescopeStartupTimeout()
	return &Controller{cfg: cfg, backend: b, Gamescope: launcher}
}

func pickBackend(cfg config.Config, manager *vkms.Manager) backend.Backend {
	if strings.EqualFold(cfg.Backend, "portal") {
		return backend.NewExperimentalPortalBackend()
	}
	return backend.NewVKMSBackend(manager)
}

func (c *Controller) SessionStart(ctx context.Context) error {
	if state, err := cleanup.Load(); err == nil && gamescope.IsPIDRunning(state.GamescopePID) {
		return fmt.Errorf("session already running with pid=%d connector=%s", state.GamescopePID, state.Connector)
	}
	lock, err := cleanup.AcquireLock()
	if err != nil {
		return err
	}
	defer func() { _ = cleanup.ReleaseLock(lock) }()

	_, cfg, err := c.buildDisplayConfig()
	if err != nil {
		return err
	}
	instanceName := fmt.Sprintf("sunshine-%d", time.Now().UnixNano())
	instance, err := c.backend.Create(ctx, instanceName)
	if err != nil {
		return err
	}
	rollback := true
	defer func() {
		if rollback {
			_ = c.backend.Destroy(instance.Name)
		}
	}()

	connector, err := vkms.DiscoverConnectorWithOptions(c.cfg.ClassDRMPath, vkms.DiscoverOptions{ForceConnector: c.cfg.ForceConnector, PreferNewestVKMSConnector: c.cfg.PreferNewestVKMSConnector, Debug: c.cfg.DebugConnectorSelection})
	if err != nil {
		return err
	}
	cmd, err := c.Gamescope.Start(ctx, cfg, connector)
	if err != nil {
		return err
	}
	state := cleanup.SessionState{InstanceName: instance.Name, GamescopePID: cmd.Process.Pid, Connector: connector, Width: cfg.Width, Height: cfg.Height, FPS: cfg.RefreshHz, HDR: cfg.HDR, Backend: c.backend.Name()}
	if err := cleanup.Save(state); err != nil {
		_ = c.Gamescope.StopByPID(cmd.Process.Pid)
		return err
	}
	rollback = false
	fmt.Printf("session-start complete: backend=%s connector=%s mode=%dx%d@%d hdr=%t gamescope_pid=%d\n", c.backend.Name(), connector, cfg.Width, cfg.Height, cfg.RefreshHz, cfg.HDR, cmd.Process.Pid)
	return nil
}

func (c *Controller) SessionStop() error {
	state, err := cleanup.Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			_ = cleanup.RemoveStaleLock()
			fmt.Println("session-stop complete: state file missing, nothing to do")
			return nil
		}
		return err
	}
	if state.GamescopePID > 0 {
		_ = c.Gamescope.StopByPID(state.GamescopePID)
	}
	if state.InstanceName != "" {
		_ = c.backend.Destroy(state.InstanceName)
	}
	_ = cleanup.RemoveStaleLock()
	if err := cleanup.Remove(); err != nil {
		return err
	}
	fmt.Printf("session-stop complete: backend=%s vkms=%s gamescope_pid=%d\n", state.Backend, state.InstanceName, state.GamescopePID)
	return nil
}

func (c *Controller) Monitor(ctx context.Context) error {
	fmt.Println("monitor mode active")
	interval := c.cfg.MonitorInterval()
	maxRuntime := c.cfg.MonitorMaxRuntime()
	start := time.Now()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		if maxRuntime > 0 && time.Since(start) >= maxRuntime {
			return c.SessionStop()
		}
		if err := c.cleanupDeadSession(); err != nil {
			return err
		}
		select {
		case <-ctx.Done():
			return c.SessionStop()
		case <-ticker.C:
		}
	}
}

func (c *Controller) cleanupDeadSession() error {
	state, err := cleanup.Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if state.GamescopePID > 0 && gamescope.IsPIDRunning(state.GamescopePID) {
		return nil
	}
	return c.SessionStop()
}

func (c *Controller) Status() error {
	state, err := cleanup.Load()
	if err != nil {
		return err
	}
	fmt.Printf("backend=%s connector=%s gamescope_pid=%d running=%t mode=%dx%d@%d hdr=%t\n", state.Backend, state.Connector, state.GamescopePID, gamescope.IsPIDRunning(state.GamescopePID), state.Width, state.Height, state.FPS, state.HDR)
	return nil
}

func (c *Controller) ValidateEnv() error {
	_, err := clientdetector.Parse()
	return err
}

func (c *Controller) PrintRequest() error {
	req, err := clientdetector.Parse()
	if err != nil {
		return err
	}
	fmt.Printf("%dx%d@%d hdr=%t\n", req.Width, req.Height, req.RefreshHz, req.HDR)
	return nil
}

func (c *Controller) DetectClient() error {
	req, err := clientdetector.Parse()
	if err != nil {
		return err
	}
	fmt.Println("Detected client:")
	fmt.Println()
	fmt.Printf("Resolution: %dx%d\n", req.Width, req.Height)
	fmt.Printf("Refresh: %dhz\n", req.RefreshHz)
	fmt.Printf("Aspect ratio: %.4g\n", req.AspectRatio)
	if req.ClientName != "" {
		fmt.Printf("Client name: %s\n", req.ClientName)
	}
	return nil
}

func (c *Controller) ShowConfig() error {
	req, cfg, err := c.buildDisplayConfig()
	if err != nil {
		return err
	}
	fmt.Println("Detected client:")
	fmt.Println()
	fmt.Printf("Resolution: %dx%d\n", req.Width, req.Height)
	fmt.Printf("Refresh: %dhz\n", req.RefreshHz)
	fmt.Printf("Aspect ratio: %.4g\n", req.AspectRatio)
	fmt.Println()
	fmt.Println("Generated display config:")
	fmt.Println()
	fmt.Printf("Resolution: %dx%d\n", cfg.Width, cfg.Height)
	fmt.Printf("Refresh: %dhz\n", cfg.RefreshHz)
	fmt.Printf("HDR: %t\n", cfg.HDR)
	if len(cfg.GamescopeFlags) > 0 {
		fmt.Printf("Gamescope flags: %s\n", strings.Join(cfg.GamescopeFlags, " "))
	}
	fmt.Printf("Disable physical monitors: %t\n", cfg.DisablePhysicalMonitors)
	return nil
}

func (c *Controller) ShowRules() error {
	loadedRules, err := rules.LoadDefault()
	if err != nil {
		return err
	}
	if loadedRules.IsEmpty() {
		fmt.Printf("No rules loaded (optional file missing or empty): %s\n", rules.DefaultPath())
		return nil
	}
	output, err := yaml.Marshal(loadedRules)
	if err != nil {
		return err
	}
	fmt.Printf("Rules loaded from %s:\n\n%s", rules.DefaultPath(), string(output))
	return nil
}

func (c *Controller) buildDisplayConfig() (clientdetector.ClientRequest, display.DisplayConfig, error) {
	req, err := clientdetector.Parse()
	if err != nil {
		return clientdetector.ClientRequest{}, display.DisplayConfig{}, err
	}
	base := display.FromClientRequest(req)
	loadedRules, err := rules.LoadDefault()
	if err != nil {
		return clientdetector.ClientRequest{}, display.DisplayConfig{}, err
	}
	finalCfg, _, err := rules.Apply(loadedRules, req, base)
	if err != nil {
		return clientdetector.ClientRequest{}, display.DisplayConfig{}, err
	}
	return req, finalCfg, nil
}

func (c *Controller) CleanupStale() error {
	_ = cleanup.RemoveStaleLock()
	state, err := cleanup.Load()
	if err == nil && !gamescope.IsPIDRunning(state.GamescopePID) {
		return cleanup.Remove()
	}
	return nil
}

func (c *Controller) Doctor() error {
	bins := []string{"gamescope", "modprobe"}
	for _, b := range bins {
		if _, err := exec.LookPath(b); err != nil {
			return fmt.Errorf("missing required binary: %s", b)
		}
	}
	if _, err := os.Stat("/sys/class/drm"); err != nil {
		return fmt.Errorf("drm sysfs not available: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(cleanup.StateFilePath()), 0o755); err != nil {
		return fmt.Errorf("runtime path not writable: %w", err)
	}
	fmt.Println("doctor: OK")
	return nil
}
