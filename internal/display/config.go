package display

import "github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/clientdetector"

// DisplayConfig is the final configuration used to create and stream the virtual display.
type DisplayConfig struct {
	Width                   int
	Height                  int
	RefreshHz               int
	HDR                     bool
	GamescopeFlags          []string
	DisablePhysicalMonitors bool
}

func FromClientRequest(req clientdetector.ClientRequest) DisplayConfig {
	return DisplayConfig{Width: req.Width, Height: req.Height, RefreshHz: req.RefreshHz, HDR: req.HDR}
}
