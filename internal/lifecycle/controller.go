package lifecycle

import (
	"context"
	"fmt"
	"time"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/cleanup"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/clientdetector"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/gamescope"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/vkms"
)

type Controller struct {
	VKMS      *vkms.Manager
	Gamescope *gamescope.Launcher
}

func NewController() *Controller {
	return &Controller{VKMS: vkms.NewManager(), Gamescope: gamescope.NewLauncher()}
}

func (c *Controller) SessionStart(ctx context.Context) error {
	req, err := clientdetector.ParseFromEnv()
	if err != nil {
		return err
	}

	instanceName := fmt.Sprintf("sunshine-%d", time.Now().UnixNano())
	instance, err := c.VKMS.Create(ctx, instanceName, req)
	if err != nil {
		return err
	}

	cmd, err := c.Gamescope.Start(ctx, req)
	if err != nil {
		_ = c.VKMS.Destroy(instance.Name)
		return err
	}

	state := cleanup.SessionState{
		InstanceName: instance.Name,
		GamescopePID: cmd.Process.Pid,
		Width:        req.Width,
		Height:       req.Height,
		FPS:          req.RefreshRate,
		HDR:          req.HDR,
	}

	if err := cleanup.Save(state); err != nil {
		_ = gamescope.StopByPID(cmd.Process.Pid)
		_ = c.VKMS.Destroy(instance.Name)
		return err
	}

	fmt.Printf("session-start complete: vkms=%s mode=%dx%d@%d hdr=%t gamescope_pid=%d\n",
		instance.Path, req.Width, req.Height, req.RefreshRate, req.HDR, cmd.Process.Pid)
	return nil
}

func (c *Controller) SessionStop() error {
	state, err := cleanup.Load()
	if err != nil {
		return err
	}

	if state.GamescopePID > 0 {
		_ = gamescope.StopByPID(state.GamescopePID)
	}

	if err := c.VKMS.Destroy(state.InstanceName); err != nil {
		return err
	}

	if err := cleanup.Remove(); err != nil {
		return err
	}

	fmt.Printf("session-stop complete: vkms=%s gamescope_pid=%d\n", state.InstanceName, state.GamescopePID)
	return nil
}

func (c *Controller) Monitor(ctx context.Context) error {
	fmt.Println("monitor mode active")
	<-ctx.Done()
	_ = c.SessionStop()
	return nil
}
