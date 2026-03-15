package rules

import (
	"fmt"
	"strings"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/clientdetector"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/display"
)

type ApplyResult struct {
	MatchedRules int
}

func Apply(cfg Config, client clientdetector.ClientRequest, base display.DisplayConfig) (display.DisplayConfig, ApplyResult, error) {
	result := base
	applied := 0
	for _, rule := range cfg.Rules {
		match, err := rule.Match.Matches(client)
		if err != nil {
			return display.DisplayConfig{}, ApplyResult{}, err
		}
		if !match {
			continue
		}
		applyOverride(&result, rule.Override)
		applied++
	}
	return result, ApplyResult{MatchedRules: applied}, nil
}

func (m MatchCriteria) Matches(client clientdetector.ClientRequest) (bool, error) {
	if m.Width != nil && client.Width != *m.Width {
		return false, nil
	}
	if m.Height != nil && client.Height != *m.Height {
		return false, nil
	}
	if m.RefreshHz != nil && client.RefreshHz != *m.RefreshHz {
		return false, nil
	}
	if name := strings.TrimSpace(m.ClientName); name != "" && !strings.EqualFold(name, client.ClientName) {
		return false, nil
	}
	if strings.TrimSpace(m.AspectRatio) != "" {
		ruleRatio, err := parseAspectRatio(m.AspectRatio)
		if err != nil {
			return false, err
		}
		if ruleRatio != client.AspectRatio {
			return false, nil
		}
	}
	return true, nil
}

func applyOverride(cfg *display.DisplayConfig, override Override) {
	if override.Width != nil {
		cfg.Width = *override.Width
	}
	if override.Height != nil {
		cfg.Height = *override.Height
	}
	if override.RefreshHz != nil {
		cfg.RefreshHz = *override.RefreshHz
	}
	if override.HDR != nil {
		cfg.HDR = *override.HDR
	}
	if override.DisablePhysicalMonitors != nil {
		cfg.DisablePhysicalMonitors = *override.DisablePhysicalMonitors
	}
	if len(override.GamescopeFlags) > 0 {
		cfg.GamescopeFlags = append(cfg.GamescopeFlags, override.GamescopeFlags...)
	}
}

func parseAspectRatio(raw string) (float64, error) {
	parts := strings.Split(strings.TrimSpace(raw), ":")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid aspect ratio %q; expected format W:H", raw)
	}
	left := strings.TrimSpace(parts[0])
	right := strings.TrimSpace(parts[1])
	if left == "" || right == "" {
		return 0, fmt.Errorf("invalid aspect ratio %q; expected format W:H", raw)
	}
	var w, h int
	if _, err := fmt.Sscanf(left+":"+right, "%d:%d", &w, &h); err != nil || w <= 0 || h <= 0 {
		return 0, fmt.Errorf("invalid aspect ratio %q; expected positive integers", raw)
	}
	tmp := clientdetector.ClientRequest{Width: w, Height: h}
	tmp.CalculateAspectRatio()
	return tmp.AspectRatio, nil
}
