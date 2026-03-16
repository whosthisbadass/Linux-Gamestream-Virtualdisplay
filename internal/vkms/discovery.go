package vkms

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type DiscoverOptions struct {
	ForceConnector            string
	PreferNewestVKMSConnector bool
	Debug                     bool
	DebugOut                  io.Writer
}

func DiscoverConnector(classDRMPath string) (string, error) {
	return DiscoverConnectorWithOptions(classDRMPath, DiscoverOptions{ForceConnector: strings.TrimSpace(os.Getenv("SVD_FORCE_CONNECTOR"))})
}

func DiscoverConnectorWithOptions(classDRMPath string, opts DiscoverOptions) (string, error) {
	if isTruthy(os.Getenv("SVD_DRY_RUN")) && opts.ForceConnector != "" {
		return opts.ForceConnector, nil
	}
	entries, err := os.ReadDir(classDRMPath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", classDRMPath, err)
	}
	var candidates []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "card") || !strings.Contains(name, "-") {
			continue
		}
		status, err := os.ReadFile(filepath.Join(classDRMPath, name, "status"))
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(status)) == "connected" {
			candidates = append(candidates, name)
		}
	}
	sort.Strings(candidates)
	debugf(opts, "connected candidates: %v", candidates)
	if len(candidates) == 0 {
		if opts.ForceConnector != "" {
			return opts.ForceConnector, nil
		}
		return "", fmt.Errorf("no connected DRM connector found under %s", classDRMPath)
	}
	if opts.ForceConnector != "" {
		debugf(opts, "forced connector selected: %s", opts.ForceConnector)
		return opts.ForceConnector, nil
	}
	if opts.PreferNewestVKMSConnector {
		for i := len(candidates) - 1; i >= 0; i-- {
			if strings.Contains(strings.ToUpper(candidates[i]), "VIRTUAL") {
				debugf(opts, "prefer newest virtual connector: %s", candidates[i])
				return candidates[i], nil
			}
		}
		return "", fmt.Errorf("no VIRTUAL DRM connector found among [%s]; set SVD_FORCE_CONNECTOR or disable SVD_PREFER_NEWEST_VKMS_CONNECTOR", strings.Join(candidates, ", "))
	}
	if len(candidates) == 1 {
		return candidates[0], nil
	}
	for _, c := range candidates {
		if strings.Contains(strings.ToUpper(c), "VIRTUAL") {
			debugf(opts, "selected virtual connector: %s", c)
			return c, nil
		}
	}
	return "", fmt.Errorf("multiple connected DRM connectors found (%s); set SVD_FORCE_CONNECTOR", strings.Join(candidates, ", "))
}

func debugf(opts DiscoverOptions, format string, args ...any) {
	if !opts.Debug {
		return
	}
	out := opts.DebugOut
	if out == nil {
		out = os.Stderr
	}
	_, _ = fmt.Fprintf(out, "connector-discovery: "+format+"\n", args...)
}
