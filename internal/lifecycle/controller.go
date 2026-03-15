package lifecycle

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/cleanup"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/clientdetector"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/gamescope"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/vkms"
)

type Controller struct {
	VKMS         *vkms.Manager
	Gamescope    *gamescope.Launcher
	ClassDRMPath string
}

func NewController() *Controller {
	return &Controller{VKMS: vkms.NewManager(), Gamescope: gamescope.NewLauncher(), ClassDRMPath: "/sys/class/drm"}
}

func (c *Controller) SessionStart(ctx context.Context) error {
	req, err := clientdetector.ParseFromEnv()
	if err != nil {
		return err
	}

	instanceName := fmt.Sprintf("sunshine-%d", time.Now().UnixNano())
	instance, err := c.VKMS.Create(ctx, instanceName)
	if err != nil {
		return err
	}

	connector, err := vkms.DiscoverConnector(c.ClassDRMPath)
	if err != nil {
		_ = c.VKMS.Destroy(instance.Name)
		return err
	}

	cmd, err := c.Gamescope.Start(ctx, req, connector)
	if err != nil {
		_ = c.VKMS.Destroy(instance.Name)
		return err
	}

	state := cleanup.SessionState{
		InstanceName: instance.Name,
		GamescopePID: cmd.Process.Pid,
		Connector:    connector,
		Width:        req.Width,
		Height:       req.Height,
		FPS:          req.RefreshRate,
		HDR:          req.HDR,
	}

	if err := cleanup.Save(state); err != nil {
		_ = c.Gamescope.StopByPID(cmd.Process.Pid)
		_ = c.VKMS.Destroy(instance.Name)
		return err
	}

	fmt.Printf("session-start complete: vkms=%s connector=%s mode=%dx%d@%d hdr=%t gamescope_pid=%d\n",
		instance.Path, connector, req.Width, req.Height, req.RefreshRate, req.HDR, cmd.Process.Pid)
	return nil
}

func (c *Controller) SessionStop() error {
	state, err := cleanup.Load()
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Println("session-stop complete: state file missing, nothing to do")
			return nil
		}
		return err
	}

	if state.GamescopePID > 0 {
		_ = c.Gamescope.StopByPID(state.GamescopePID)
	}

	if state.InstanceName != "" {
		if err := c.VKMS.Destroy(state.InstanceName); err != nil {
			return err
		}
	}

	if err := cleanup.Remove(); err != nil {
		return err
	}

	fmt.Printf("session-stop complete: vkms=%s gamescope_pid=%d\n", state.InstanceName, state.GamescopePID)
	return nil
}

func (c *Controller) Monitor(ctx context.Context) error {
	fmt.Println("monitor mode active")
	interval := envDurationSeconds("SVD_MONITOR_INTERVAL_SEC", 5)
	maxRuntime := envDurationSeconds("SVD_MONITOR_MAX_RUNTIME_SEC", 0)

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

func envDurationSeconds(key string, fallback int) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return time.Duration(fallback) * time.Second
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return time.Duration(fallback) * time.Second
	}
	return time.Duration(value) * time.Second
}
