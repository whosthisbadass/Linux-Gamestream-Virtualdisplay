package vkms

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func DiscoverConnector(classDRMPath string) (string, error) {
	forced := strings.TrimSpace(os.Getenv("SVD_FORCE_CONNECTOR"))
	if isTruthy(os.Getenv("SVD_DRY_RUN")) && forced != "" {
		return forced, nil
	}

	entries, err := os.ReadDir(classDRMPath)
	if err != nil {
		return "", fmt.Errorf("read %s: %w", classDRMPath, err)
	}

	candidates := make([]string, 0)
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "card") || !strings.Contains(name, "-") {
			continue
		}
		statusPath := filepath.Join(classDRMPath, name, "status")
		status, err := os.ReadFile(statusPath)
		if err != nil {
			continue
		}
		if strings.TrimSpace(string(status)) == "connected" {
			candidates = append(candidates, name)
		}
	}

	sort.Strings(candidates)
	if len(candidates) == 0 {
		if forced != "" {
			return forced, nil
		}
		return "", fmt.Errorf("no connected DRM connector found under %s", classDRMPath)
	}
	if forced != "" {
		return forced, nil
	}
	if len(candidates) > 1 {
		return "", fmt.Errorf("multiple connected DRM connectors found (%s); set SVD_FORCE_CONNECTOR", strings.Join(candidates, ", "))
	}
	return candidates[0], nil
}
