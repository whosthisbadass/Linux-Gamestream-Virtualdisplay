package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Backend                    string `json:"backend"`
	ClassDRMPath               string `json:"class_drm_path"`
	ForceConnector             string `json:"force_connector,omitempty"`
	PreferNewestVKMSConnector  bool   `json:"prefer_newest_vkms_connector"`
	DebugConnectorSelection    bool   `json:"debug_connector_selection"`
	DryRun                     bool   `json:"dry_run"`
	GamescopeTarget            string `json:"gamescope_target"`
	GamescopeGenerateDRMMode   string `json:"gamescope_generate_drm_mode"`
	GamescopeLogPath           string `json:"gamescope_log_path"`
	GamescopeStartupTimeoutSec int    `json:"gamescope_startup_timeout_sec"`
	MonitorIntervalSec         int    `json:"monitor_interval_sec"`
	MonitorMaxRuntimeSec       int    `json:"monitor_max_runtime_sec"`
}

func Default() Config {
	return Config{
		Backend:                    "vkms",
		ClassDRMPath:               "/sys/class/drm",
		GamescopeTarget:            "sleep infinity",
		GamescopeGenerateDRMMode:   "cvt",
		GamescopeLogPath:           filepath.Join(runtimeDir(), "sunshine-virtual-display", "gamescope.log"),
		GamescopeStartupTimeoutSec: 10,
		MonitorIntervalSec:         5,
		MonitorMaxRuntimeSec:       0,
	}
}

func Load() (Config, error) {
	cfg := Default()
	if err := applyConfigFile(&cfg); err != nil {
		return Config{}, err
	}
	applyEnv(&cfg)
	return cfg, cfg.Validate()
}

func (c Config) Validate() error {
	if c.Backend == "" {
		return fmt.Errorf("backend must not be empty")
	}
	if c.ClassDRMPath == "" {
		return fmt.Errorf("class drm path must not be empty")
	}
	if c.MonitorIntervalSec < 0 || c.MonitorMaxRuntimeSec < 0 || c.GamescopeStartupTimeoutSec <= 0 {
		return fmt.Errorf("timeout values must be positive")
	}
	validModes := map[string]bool{"cvt": true, "gtf": true, "": true, "0": true, "off": true, "false": true}
	if !validModes[strings.ToLower(strings.TrimSpace(c.GamescopeGenerateDRMMode))] {
		return fmt.Errorf("gamescope_generate_drm_mode %q is invalid; valid values: cvt, gtf, off", c.GamescopeGenerateDRMMode)
	}
	return nil
}

func (c Config) MonitorInterval() time.Duration {
	return time.Duration(c.MonitorIntervalSec) * time.Second
}
func (c Config) MonitorMaxRuntime() time.Duration {
	return time.Duration(c.MonitorMaxRuntimeSec) * time.Second
}
func (c Config) GamescopeStartupTimeout() time.Duration {
	return time.Duration(c.GamescopeStartupTimeoutSec) * time.Second
}

func (c Config) Dump() (string, error) {
	out, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func applyConfigFile(cfg *Config) error {
	path := strings.TrimSpace(os.Getenv("SVD_CONFIG_FILE"))
	if path == "" {
		path = "/etc/sunshine-virtual-display/sunshine-virtual-display.json"
	}
	payload, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("read config file: %w", err)
	}
	if err := json.Unmarshal(payload, cfg); err != nil {
		return fmt.Errorf("parse config file: %w", err)
	}
	return nil
}

func applyEnv(cfg *Config) {
	cfg.Backend = envString("SVD_BACKEND", cfg.Backend)
	cfg.ClassDRMPath = envString("SVD_CLASS_DRM_PATH", cfg.ClassDRMPath)
	cfg.ForceConnector = envString("SVD_FORCE_CONNECTOR", cfg.ForceConnector)
	cfg.PreferNewestVKMSConnector = envBool("SVD_PREFER_NEWEST_VKMS_CONNECTOR", cfg.PreferNewestVKMSConnector)
	cfg.DebugConnectorSelection = envBool("SVD_DEBUG_CONNECTOR", cfg.DebugConnectorSelection)
	cfg.DryRun = envBool("SVD_DRY_RUN", cfg.DryRun)
	cfg.GamescopeTarget = envString("SUNSHINE_VD_GAMESCOPE_TARGET", cfg.GamescopeTarget)
	cfg.GamescopeGenerateDRMMode = envString("SVD_GAMESCOPE_GENERATE_DRM_MODE", cfg.GamescopeGenerateDRMMode)
	cfg.GamescopeLogPath = envString("SVD_GAMESCOPE_LOG_PATH", cfg.GamescopeLogPath)
	cfg.GamescopeStartupTimeoutSec = envInt("SVD_GAMESCOPE_STARTUP_TIMEOUT_SEC", cfg.GamescopeStartupTimeoutSec)
	cfg.MonitorIntervalSec = envInt("SVD_MONITOR_INTERVAL_SEC", cfg.MonitorIntervalSec)
	cfg.MonitorMaxRuntimeSec = envInt("SVD_MONITOR_MAX_RUNTIME_SEC", cfg.MonitorMaxRuntimeSec)
}

func runtimeDir() string {
	if d := strings.TrimSpace(os.Getenv("XDG_RUNTIME_DIR")); d != "" {
		return d
	}
	return "/tmp"
}

func envString(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	raw := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	switch raw {
	case "1", "true", "yes", "on":
		return true
	case "0", "false", "no", "off":
		return false
	case "":
		return fallback
	default:
		return fallback
	}
}

func envInt(key string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	if n, err := strconv.Atoi(raw); err == nil {
		return n
	}
	return fallback
}
