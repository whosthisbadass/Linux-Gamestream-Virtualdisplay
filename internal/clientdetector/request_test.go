package clientdetector

import "testing"

func TestParse(t *testing.T) {
	t.Setenv("SUNSHINE_CLIENT_WIDTH", "2560")
	t.Setenv("SUNSHINE_CLIENT_HEIGHT", "1600")
	t.Setenv("SUNSHINE_CLIENT_FPS", "120")
	t.Setenv("SUNSHINE_CLIENT_HDR", "1")
	t.Setenv("SUNSHINE_CLIENT_NAME", "Steam Deck")

	req, err := Parse()
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}

	if req.Width != 2560 || req.Height != 1600 || req.RefreshHz != 120 || !req.HDR {
		t.Fatalf("unexpected request: %+v", req)
	}
	if req.AspectRatio != 1.6 {
		t.Fatalf("unexpected aspect ratio: %f", req.AspectRatio)
	}
	if !req.IsSteamDeck() {
		t.Fatalf("expected steam deck detection")
	}
}

func TestParseFromEnvCompatibility(t *testing.T) {
	t.Setenv("SUNSHINE_CLIENT_WIDTH", "1920")
	t.Setenv("SUNSHINE_CLIENT_HEIGHT", "1080")
	t.Setenv("SUNSHINE_CLIENT_FPS", "60")

	req, err := ParseFromEnv()
	if err != nil {
		t.Fatalf("ParseFromEnv returned error: %v", err)
	}
	if req.Width != 1920 || req.Height != 1080 || req.RefreshRate != 60 {
		t.Fatalf("unexpected compatibility request: %+v", req)
	}
}
