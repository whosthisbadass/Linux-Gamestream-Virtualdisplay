package clientdetector

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

const (
	minWidth  = 320
	maxWidth  = 7680
	minHeight = 240
	maxHeight = 4320
	minFPS    = 1
	maxFPS    = 480
)

// ClientRequest describes a normalized Sunshine client display request.
type ClientRequest struct {
	Width       int
	Height      int
	RefreshHz   int
	AspectRatio float64
	ClientName  string
	HDR         bool
}

// ClientDisplayRequest is maintained as a compatibility adapter for existing call sites.
type ClientDisplayRequest struct {
	Width       int
	Height      int
	RefreshRate int
	HDR         bool
}

// Parse reads Sunshine-provided environment variables and normalizes the request.
func Parse() (ClientRequest, error) {
	width, err := requiredPositiveInt("SUNSHINE_CLIENT_WIDTH")
	if err != nil {
		return ClientRequest{}, err
	}
	if width < minWidth || width > maxWidth {
		return ClientRequest{}, fmt.Errorf("SUNSHINE_CLIENT_WIDTH=%d is out of supported range [%d, %d]", width, minWidth, maxWidth)
	}

	height, err := requiredPositiveInt("SUNSHINE_CLIENT_HEIGHT")
	if err != nil {
		return ClientRequest{}, err
	}
	if height < minHeight || height > maxHeight {
		return ClientRequest{}, fmt.Errorf("SUNSHINE_CLIENT_HEIGHT=%d is out of supported range [%d, %d]", height, minHeight, maxHeight)
	}

	fps, err := requiredPositiveInt("SUNSHINE_CLIENT_FPS")
	if err != nil {
		return ClientRequest{}, err
	}
	if fps < minFPS || fps > maxFPS {
		return ClientRequest{}, fmt.Errorf("SUNSHINE_CLIENT_FPS=%d is out of supported range [%d, %d]", fps, minFPS, maxFPS)
	}

	hdr, err := optionalBool("SUNSHINE_CLIENT_HDR", false)
	if err != nil {
		return ClientRequest{}, err
	}

	req := ClientRequest{
		Width:      width,
		Height:     height,
		RefreshHz:  fps,
		ClientName: strings.TrimSpace(os.Getenv("SUNSHINE_CLIENT_NAME")),
		HDR:        hdr,
	}
	req.CalculateAspectRatio()
	return req, nil
}

// ParseFromEnv reads the request and maps it to the legacy display request shape.
func ParseFromEnv() (ClientDisplayRequest, error) {
	req, err := Parse()
	if err != nil {
		return ClientDisplayRequest{}, err
	}
	return ClientDisplayRequest{Width: req.Width, Height: req.Height, RefreshRate: req.RefreshHz, HDR: req.HDR}, nil
}

func (r *ClientRequest) CalculateAspectRatio() {
	if r.Height <= 0 {
		r.AspectRatio = 0
		return
	}
	r.AspectRatio = roundTo(float64(r.Width)/float64(r.Height), 4)
}

func (r ClientRequest) IsUltrawide() bool { return r.AspectRatio >= 2.0 }
func (r ClientRequest) IsPortrait() bool  { return r.Height > r.Width }
func (r ClientRequest) IsSteamDeck() bool {
	name := strings.ToLower(strings.TrimSpace(r.ClientName))
	return strings.Contains(name, "steam deck") || strings.Contains(name, "steamdeck")
}

func roundTo(v float64, places int) float64 {
	if places <= 0 {
		return math.Round(v)
	}
	multiplier := math.Pow10(places)
	return math.Round(v*multiplier) / multiplier
}

func requiredPositiveInt(key string) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return 0, fmt.Errorf("required environment variable %s is missing", key)
	}

	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("environment variable %s must be a positive integer, got %q", key, raw)
	}

	return value, nil
}

func optionalBool(key string, fallback bool) (bool, error) {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if raw == "" {
		return fallback, nil
	}
	switch raw {
	case "1", "true", "yes", "on":
		return true, nil
	case "0", "false", "no", "off":
		return false, nil
	default:
		return false, fmt.Errorf("environment variable %s must be a boolean value, got %q", key, raw)
	}
}
