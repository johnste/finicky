//go:build darwin

package version

import (
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// getCurrentVersionPlatform reads CFBundleVersion from the app bundle's
// Info.plist via the macOS `defaults` tool. Fallback only — the
// ldflags-injected version wins.
func getCurrentVersionPlatform() string {
	// Get the bundle path
	bundlePath := os.Getenv("BUNDLE_PATH")
	if bundlePath == "" {
		execPath, err := os.Executable()
		if err != nil {
			slog.Error("Error getting executable path", "error", err)
			return ""
		}

		bundlePath = filepath.Join(filepath.Dir(execPath), "..", "Info.plist")
	}

	// Read and parse Info.plist
	cmd := exec.Command("defaults", "read", bundlePath, "CFBundleVersion")
	output, err := cmd.Output()
	if err != nil {
		slog.Error("Error reading version from Info.plist", "error", err)
		return ""
	}

	version := strings.TrimSpace(string(output))

	if version == "" {
		slog.Error("Could not determine current version")
		return "dev"
	}

	return version
}
