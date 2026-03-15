package gamescope

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/clientdetector"
)

type Launcher struct {
	LogPath     string
	StopTimeout time.Duration
}

func NewLauncher() *Launcher {
	return &Launcher{LogPath: "/tmp/sunshine-virtual-display-gamescope.log", StopTimeout: 5 * time.Second}
}

func (l *Launcher) BuildArgs(req clientdetector.ClientDisplayRequest, connector string) ([]string, error) {
	connector = strings.TrimSpace(connector)
	if connector == "" {
		return nil, fmt.Errorf("gamescope connector is required")
	}

	args := []string{
		"-O", connector,
		"-W", strconv.Itoa(req.Width),
		"-H", strconv.Itoa(req.Height),
		"-r", strconv.Itoa(req.RefreshRate),
	}

	if modeGen := strings.TrimSpace(os.Getenv("SVD_GAMESCOPE_GENERATE_DRM_MODE")); modeGen != "" {
		if modeGen != "0" && strings.ToLower(modeGen) != "off" && strings.ToLower(modeGen) != "false" {
			args = append(args, "--generate-drm-mode", modeGen)
		}
	} else {
		args = append(args, "--generate-drm-mode", "cvt")
	}

	if req.HDR {
		args = append(args, "--hdr-enabled")
	}

	target := strings.TrimSpace(os.Getenv("SUNSHINE_VD_GAMESCOPE_TARGET"))
	if target == "" {
		target = "sleep infinity"
	}

	args = append(args, "--", "sh", "-lc", target)
	return args, nil
}

func (l *Launcher) Start(ctx context.Context, req clientdetector.ClientDisplayRequest, connector string) (*exec.Cmd, error) {
	args, err := l.BuildArgs(req, connector)
	if err != nil {
		return nil, err
	}

	cmd := exec.CommandContext(ctx, "gamescope", args...)
	logFile, err := os.OpenFile(l.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, fmt.Errorf("open gamescope log: %w", err)
	}

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

	return cmd, nil
}

func (l *Launcher) StopByPID(pid int) error {
	return StopByPID(pid, l.StopTimeout)
}

func StopByPID(pid int, timeout time.Duration) error {
	if pid <= 0 {
		return nil
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	if !isProcessRunning(proc) {
		return nil
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil && !errors.Is(err, os.ErrProcessDone) {
		return err
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if !isProcessRunning(proc) {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	if !isProcessRunning(proc) {
		return nil
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
	return isProcessRunning(proc)
}

func isProcessRunning(proc *os.Process) bool {
	err := proc.Signal(syscall.Signal(0))
	return err == nil
}
