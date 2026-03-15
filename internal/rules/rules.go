package rules

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const defaultRulesPath = "/etc/sunshine-virtual-display/rules.yaml"

type Config struct {
	Rules []Rule `yaml:"rules"`
}

type Rule struct {
	Match    MatchCriteria `yaml:"match"`
	Override Override      `yaml:"override"`
}

type MatchCriteria struct {
	Width       *int   `yaml:"width"`
	Height      *int   `yaml:"height"`
	AspectRatio string `yaml:"aspect_ratio"`
	ClientName  string `yaml:"client_name"`
	RefreshHz   *int   `yaml:"refresh"`
}

type Override struct {
	Width                   *int     `yaml:"width"`
	Height                  *int     `yaml:"height"`
	RefreshHz               *int     `yaml:"refresh"`
	GamescopeFlags          []string `yaml:"gamescope_flags"`
	HDR                     *bool    `yaml:"hdr"`
	DisablePhysicalMonitors *bool    `yaml:"disable_physical_monitors"`
}

func DefaultPath() string { return defaultRulesPath }

func LoadDefault() (Config, error) { return LoadFromPath(defaultRulesPath) }

func LoadFromPath(path string) (Config, error) {
	if path == "" {
		path = defaultRulesPath
	}
	payload, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return Config{}, nil
		}
		return Config{}, fmt.Errorf("read rules file %s: %w", path, err)
	}
	var cfg Config
	if err := yaml.Unmarshal(payload, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse rules file %s: %w", path, err)
	}
	return cfg, nil
}

func (c Config) IsEmpty() bool { return len(c.Rules) == 0 }
