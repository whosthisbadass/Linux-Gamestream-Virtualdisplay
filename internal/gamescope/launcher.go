package gamescope

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/clientdetector"
)

type Launcher struct {
	LogPath        string
	StopTimeout    time.Duration
	StartupTimeout time.Duration
	TargetCommand  string
	ModeGeneration string
}

func NewLauncher() *Launcher {
	return &Launcher{LogPath: "/tmp/sunshine-virtual-display-gamescope.log", StopTimeout: 5 * time.Second, StartupTimeout: 10 * time.Second, TargetCommand: "sleep infinity", ModeGeneration: "cvt"}
}

func (l *Launcher) BuildArgs(req clientdetector.ClientDisplayRequest, connector string) ([]string, error) {
	connector = strings.TrimSpace(connector)
	if connector == "" {
		return nil, fmt.Errorf("gamescope connector is required")
	}

	args := []string{"-O", connector, "-W", strconv.Itoa(req.Width), "-H", strconv.Itoa(req.Height), "-r", strconv.Itoa(req.RefreshRate)}
	if modeGen := strings.TrimSpace(l.ModeGeneration); modeGen != "" && modeGen != "0" && strings.ToLower(modeGen) != "off" && strings.ToLower(modeGen) != "false" {
		args = append(args, "--generate-drm-mode", modeGen)
	}
	if req.HDR {
		args = append(args, "--hdr-enabled")
	}
	target := strings.TrimSpace(l.TargetCommand)
	if target == "" {
		target = "sleep infinity"
	}
	return append(args, "--", "sh", "-lc", target), nil
}

func (l *Launcher) Start(ctx context.Context, req clientdetector.ClientDisplayRequest, connector string) (*exec.Cmd, error) {
	args, err := l.BuildArgs(req, connector)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(l.LogPath), 0o755); err != nil {
		return nil, fmt.Errorf("create gamescope log dir: %w", err)
	}
	logFile, err := os.OpenFile(l.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open gamescope log: %w", err)
	}
	cmd := exec.CommandContext(ctx, "gamescope", args...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return nil, fmt.Errorf("start gamescope: %w", err)
	}
	go func() {
		_ = cmd.Wait()
		_ = logFile.Close()
	}()

	deadline := time.Now().Add(l.StartupTimeout)
	for time.Now().Before(deadline) {
		if !IsPIDRunning(cmd.Process.Pid) {
			return nil, fmt.Errorf("gamescope exited before becoming ready; check log at %s", l.LogPath)
		}
		time.Sleep(100 * time.Millisecond)
	}
	return cmd, nil
}

func (l *Launcher) StopByPID(pid int) error { return StopByPID(pid, l.StopTimeout) }

func StopByPID(pid int, timeout time.Duration) error {
	if pid <= 0 {
		return nil
	}
	if !IsPIDRunning(pid) {
		return nil
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	if err := proc.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !IsPIDRunning(pid) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := proc.Signal(syscall.SIGKILL); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}
	return nil
}

func IsPIDRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	return proc.Signal(syscall.Signal(0)) == nil
}
