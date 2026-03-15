package gamescope

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/clientdetector"
)

type Launcher struct {
	LogPath string
}

func NewLauncher() *Launcher {
	return &Launcher{LogPath: "/tmp/sunshine-virtual-display-gamescope.log"}
}

func (l *Launcher) Start(ctx context.Context, req clientdetector.ClientDisplayRequest) (*exec.Cmd, error) {
	target := strings.TrimSpace(os.Getenv("SUNSHINE_VD_GAMESCOPE_TARGET"))
	if target == "" {
		target = "sleep infinity"
	}

	cmd := exec.CommandContext(ctx, "gamescope",
		"-W", fmt.Sprintf("%d", req.Width),
		"-H", fmt.Sprintf("%d", req.Height),
		"-r", fmt.Sprintf("%d", req.RefreshRate),
		"--",
		"sh", "-lc", target,
	)

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

func StopByPID(pid int) error {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}
	return proc.Kill()
}
