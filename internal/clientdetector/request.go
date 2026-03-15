package clientdetector

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// ClientDisplayRequest is a direct representation of Sunshine's client-provided
// display request. No negotiation or approximation is performed.
type ClientDisplayRequest struct {
	Width       int
	Height      int
	RefreshRate int
	HDR         bool
}

// ParseFromEnv reads the required Sunshine-provided environment variables.
func ParseFromEnv() (ClientDisplayRequest, error) {
	width, err := requiredPositiveInt("SUNSHINE_CLIENT_WIDTH")
	if err != nil {
		return ClientDisplayRequest{}, err
	}

	height, err := requiredPositiveInt("SUNSHINE_CLIENT_HEIGHT")
	if err != nil {
		return ClientDisplayRequest{}, err
	}

	fps, err := requiredPositiveInt("SUNSHINE_CLIENT_FPS")
	if err != nil {
		return ClientDisplayRequest{}, err
	}

	hdr, err := requiredBool("SUNSHINE_CLIENT_HDR")
	if err != nil {
		return ClientDisplayRequest{}, err
	}

	return ClientDisplayRequest{
		Width:       width,
		Height:      height,
		RefreshRate: fps,
		HDR:         hdr,
	}, nil
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

func requiredBool(key string) (bool, error) {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(key)))
	if raw == "" {
		return false, fmt.Errorf("required environment variable %s is missing", key)
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
