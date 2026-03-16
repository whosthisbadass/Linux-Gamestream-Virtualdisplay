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

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/display"
)

type Launcher struct {
	LogPath        string
	StopTimeout    time.Duration
	StartupTimeout time.Duration
	TargetCommand  string
	ModeGeneration string
}

const pollInterval = 100 * time.Millisecond

func NewLauncher() *Launcher {
	return &Launcher{
		LogPath:        "/tmp/sunshine-virtual-display-gamescope.log",
		StopTimeout:    5 * time.Second,
		StartupTimeout: 10 * time.Second,
		TargetCommand:  "sleep infinity",
		ModeGeneration: "cvt",
	}
}

func (l *Launcher) BuildArgs(cfg display.DisplayConfig, connector string) ([]string, error) {
	connector = strings.TrimSpace(connector)
	if connector == "" {
		return nil, fmt.Errorf("gamescope connector is required")
	}

	args := []string{"-O", connector, "-W", strconv.Itoa(cfg.Width), "-H", strconv.Itoa(cfg.Height), "-r", strconv.Itoa(cfg.RefreshHz)}
	if modeGen := strings.TrimSpace(l.ModeGeneration); modeGen != "" && modeGen != "0" && strings.ToLower(modeGen) != "off" && strings.ToLower(modeGen) != "false" {
		args = append(args, "--generate-drm-mode", modeGen)
	}
	if cfg.HDR {
		args = append(args, "--hdr-enabled")
	}
	if len(cfg.GamescopeFlags) > 0 {
		args = append(args, cfg.GamescopeFlags...)
	}
	target := strings.TrimSpace(l.TargetCommand)
	if target == "" {
		target = "sleep infinity"
	}
	if err := validateTargetCommand(target); err != nil {
		return nil, fmt.Errorf("gamescope target command: %w", err)
	}
	return append(args, "--", "sh", "-lc", target), nil
}

// validateTargetCommand rejects target commands containing shell metacharacters
// that could be used for injection. Use a wrapper script for commands that
// require shell features like pipes, redirections, or variable expansion.
func validateTargetCommand(cmd string) error {
	for _, r := range cmd {
		switch r {
		case ';', '&', '|', '`', '$', '(', ')', '{', '}', '<', '>', '!', '\\', '\n', '\r', '\x00', '#':
			return fmt.Errorf("forbidden character %q; use a wrapper script for complex commands", string(r))
		}
	}
	return nil
}

func (l *Launcher) Start(ctx context.Context, cfg display.DisplayConfig, connector string) (*exec.Cmd, error) {
	args, err := l.BuildArgs(cfg, connector)
	if err != nil {
		return nil, err
	}
	logFile, err := l.openLogFile()
	if err != nil {
		return nil, err
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

	// Poll until Gamescope has been alive for a short stability window, or bail
	// early if it crashes. StartupTimeout is the maximum time to wait; once the
	// process has been running for stabilityWindow without crashing, it is
	// considered ready and we return immediately rather than always blocking for
	// the full timeout.
	const stabilityWindow = 500 * time.Millisecond
	deadline := time.Now().Add(l.StartupTimeout)
	stable := time.Now().Add(stabilityWindow)
	for time.Now().Before(deadline) {
		if !IsPIDRunning(cmd.Process.Pid) {
			return nil, fmt.Errorf("gamescope exited before becoming ready; check log at %s", l.LogPath)
		}
		if !time.Now().Before(stable) {
			return cmd, nil
		}
		time.Sleep(pollInterval)
	}
	return cmd, nil
}

func (l *Launcher) openLogFile() (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(l.LogPath), 0o755); err != nil {
		return nil, fmt.Errorf("create gamescope log dir: %w", err)
	}
	logFile, err := os.OpenFile(l.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open gamescope log: %w", err)
	}
	return logFile, nil
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
		time.Sleep(pollInterval)
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
