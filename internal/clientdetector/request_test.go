package clientdetector

import "testing"

func TestParseFromEnv(t *testing.T) {
	t.Setenv("SUNSHINE_CLIENT_WIDTH", "2560")
	t.Setenv("SUNSHINE_CLIENT_HEIGHT", "1600")
	t.Setenv("SUNSHINE_CLIENT_FPS", "120")
	t.Setenv("SUNSHINE_CLIENT_HDR", "1")

	req, err := ParseFromEnv()
	if err != nil {
		t.Fatalf("ParseFromEnv returned error: %v", err)
	}

	if req.Width != 2560 || req.Height != 1600 || req.RefreshRate != 120 || !req.HDR {
		t.Fatalf("unexpected request: %+v", req)
	}
}
