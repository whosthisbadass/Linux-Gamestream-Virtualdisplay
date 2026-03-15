package rules

import (
	"testing"

	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/clientdetector"
	"github.com/linux-gamestream-virtualdisplay/sunshine-virtual-display/internal/display"
)

func TestApplyNoRulesDynamicBehavior(t *testing.T) {
	client := clientdetector.ClientRequest{Width: 2560, Height: 1600, RefreshHz: 120}
	client.CalculateAspectRatio()

	cfg, res, err := Apply(Config{}, client, display.FromClientRequest(client))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if res.MatchedRules != 0 {
		t.Fatalf("expected no matched rules, got %d", res.MatchedRules)
	}
	if cfg.Width != 2560 || cfg.Height != 1600 || cfg.RefreshHz != 120 {
		t.Fatalf("unexpected config: %+v", cfg)
	}
}

func TestApplyRuleMatchWidth(t *testing.T) {
	refresh := 90
	client := clientdetector.ClientRequest{Width: 1280, Height: 800, RefreshHz: 60}
	client.CalculateAspectRatio()

	cfg, res, err := Apply(Config{Rules: []Rule{{Match: MatchCriteria{Width: intPtr(1280)}, Override: Override{RefreshHz: &refresh}}}}, client, display.FromClientRequest(client))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if res.MatchedRules != 1 {
		t.Fatalf("expected one match, got %d", res.MatchedRules)
	}
	if cfg.RefreshHz != 90 {
		t.Fatalf("expected refresh override to 90, got %d", cfg.RefreshHz)
	}
}

func TestApplyRuleMatchAspectRatio(t *testing.T) {
	hdr := true
	client := clientdetector.ClientRequest{Width: 2560, Height: 1600, RefreshHz: 120}
	client.CalculateAspectRatio()

	cfg, res, err := Apply(Config{Rules: []Rule{{Match: MatchCriteria{AspectRatio: "16:10"}, Override: Override{HDR: &hdr}}}}, client, display.FromClientRequest(client))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if res.MatchedRules != 1 {
		t.Fatalf("expected one match, got %d", res.MatchedRules)
	}
	if !cfg.HDR {
		t.Fatalf("expected HDR override to true")
	}
}

func TestApplyOverrideRefresh(t *testing.T) {
	refresh := 75
	client := clientdetector.ClientRequest{Width: 1920, Height: 1080, RefreshHz: 60}
	client.CalculateAspectRatio()

	cfg, _, err := Apply(Config{Rules: []Rule{{Match: MatchCriteria{RefreshHz: intPtr(60)}, Override: Override{RefreshHz: &refresh}}}}, client, display.FromClientRequest(client))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if cfg.RefreshHz != 75 {
		t.Fatalf("expected 75Hz, got %d", cfg.RefreshHz)
	}
}

func TestApplyMultipleRules(t *testing.T) {
	refresh := 90
	width := 1400
	client := clientdetector.ClientRequest{Width: 1280, Height: 800, RefreshHz: 60}
	client.CalculateAspectRatio()

	cfg, res, err := Apply(Config{Rules: []Rule{
		{Match: MatchCriteria{Width: intPtr(1280)}, Override: Override{RefreshHz: &refresh}},
		{Match: MatchCriteria{AspectRatio: "16:10"}, Override: Override{Width: &width}},
	}}, client, display.FromClientRequest(client))
	if err != nil {
		t.Fatalf("Apply returned error: %v", err)
	}
	if res.MatchedRules != 2 {
		t.Fatalf("expected two matches, got %d", res.MatchedRules)
	}
	if cfg.RefreshHz != 90 || cfg.Width != 1400 {
		t.Fatalf("unexpected merged config: %+v", cfg)
	}
}

func intPtr(v int) *int { return &v }
