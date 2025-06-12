package version

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/semver"
	"github.com/dop251/goja"
)

const updateCheckInterval = 24 * time.Hour

type ReleaseInfo struct {
	HasUpdate     bool   `json:"hasUpdate"`
	LatestVersion string `json:"latestVersion"`
	DownloadUrl   string `json:"downloadUrl"`
	ReleaseUrl    string `json:"releaseUrl"`
}

type UpdateCheckInfo struct {
	Timestamp   int64       `json:"timestamp"`
	ReleaseInfo ReleaseInfo `json:"releaseInfo"`
}

var (
	// These will be set via ldflags during build
	commitHash = "dev"
	buildDate  = "unknown"
	apiHost    = ""
)

// GetBuildInfo returns the commit hash and build date
func GetBuildInfo() (string, string) {
	return commitHash, buildDate
}

func getCacheDir() string {
	cacheDir, err := os.UserCacheDir()
	if err != nil {
		slog.Error("Error getting user cache directory", "error", err)
		return ""
	}
	cacheDir = filepath.Join(cacheDir, "Finicky")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		slog.Error("Error creating cache directory", "error", err)
		return ""
	}
	return cacheDir
}

func getLastUpdateCheck() *UpdateCheckInfo {
	cacheDir := getCacheDir()
	if cacheDir == "" {
		return nil
	}

	cacheFile := filepath.Join(cacheDir, "last_update_check.json")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil
	}

	var info UpdateCheckInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil
	}

	return &info
}

func setLastUpdateCheck(info UpdateCheckInfo) {
	cacheDir := getCacheDir()
	if cacheDir == "" {
		return
	}

	data, err := json.Marshal(info)
	if err != nil {
		slog.Error("Error marshaling update check info", "error", err)
		return
	}

	cacheFile := filepath.Join(cacheDir, "last_update_check.json")
	if err := os.WriteFile(cacheFile, data, 0644); err != nil {
		slog.Error("Error saving last update check info", "error", err)
	}
}

func GetCurrentVersion() string {
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

func checkForUpdates() (releaseInfo *ReleaseInfo) {
	currentVersion := GetCurrentVersion()
	if currentVersion == "" {
		slog.Info("Could not determine current version")
		return nil
	}

	slog.Debug("Checking update schedule...")

	updateCheckInfo := getLastUpdateCheck()
	if updateCheckInfo != nil && updateCheckInfo.Timestamp > 0 {
		timeSinceLastCheck := time.Since(time.Unix(updateCheckInfo.Timestamp, 0))
		if timeSinceLastCheck < updateCheckInterval {
			slog.Debug("Skipping update check - last checked", "duration", fmt.Sprintf("%dh %dm ago (check interval: %dh)", int(timeSinceLastCheck.Hours()), int(timeSinceLastCheck.Minutes())%60, int(updateCheckInterval.Hours())))

			updateAvailable, err := isUpdateAvailable(currentVersion, updateCheckInfo.ReleaseInfo.LatestVersion)
			if err != nil {
				slog.Error("Error checking version", "error", err)
				return nil
			}

			if updateAvailable {
				slog.Debug("Update available", "currentVersion", currentVersion, "latestVersion", updateCheckInfo.ReleaseInfo.LatestVersion)
				return &updateCheckInfo.ReleaseInfo
			} else {
				slog.Debug("Current version is up to date", "currentVersion", currentVersion, "latestVersion", updateCheckInfo.ReleaseInfo.LatestVersion)
				return nil
			}
		}
	}

	slog.Info("Checking for updates...")

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	if apiHost == "" {
		slog.Warn("apiHost is not set, won't check for updates")
		return nil
	}

	// Create request
	apiUrl := fmt.Sprintf("%s/update-check?version=%s", apiHost, currentVersion)
	req, err := http.NewRequest("GET", apiUrl, nil)
	if err != nil {
		slog.Error("Error creating request", "error", err)
		return nil
	}

	// Set User-Agent header
	req.Header.Set("User-Agent", fmt.Sprintf("finicky/%s", currentVersion))

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error making request", "error", err)
		return nil
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Error reading response", "error", err)
		return nil
	}

	// Parse releases
	if err := json.Unmarshal(body, &releaseInfo); err != nil {
		slog.Error("Error parsing release info", "error", err, "response", string(body))
		return nil
	}

	// Update the last check time
	setLastUpdateCheck(UpdateCheckInfo{
		Timestamp:   time.Now().Unix(),
		ReleaseInfo: *releaseInfo,
	})

	updateAvailable, err := isUpdateAvailable(currentVersion, releaseInfo.LatestVersion)
	if err != nil {
		slog.Error("Error checking version", "error", err)
		return nil
	}

	if updateAvailable {
		slog.Debug("Update available", "currentVersion", currentVersion, "latestVersion", releaseInfo.LatestVersion)
	} else {
		slog.Debug("Current version is up to date", "currentVersion", currentVersion, "latestVersion", releaseInfo.LatestVersion)
		return nil
	}

	return releaseInfo
}

// CheckForUpdatesIfEnabled checks if updates should be performed based on VM configuration
func CheckForUpdatesIfEnabled(vm *goja.Runtime) (releaseInfo *ReleaseInfo, updateCheckEnabled bool, err error) {

	if vm == nil {
		// Check for updates if we don't have a VM
		releaseInfo := checkForUpdates()
		return releaseInfo, true, nil
	}

	// Check checkForUpdates option
	shouldCheckForUpdates, err := vm.RunString("finickyConfigAPI.getOption('checkForUpdates', finalConfig, true)")
	if err != nil {
		return nil, true, fmt.Errorf("failed to get checkForUpdates option: %v", err)
	}

	if shouldCheckForUpdates.ToBoolean() {
		releaseInfo := checkForUpdates()
		return releaseInfo, true, nil
	} else {
		slog.Debug("Skipping update check")
	}
	return nil, false, nil
}

// isUpdateAvailable checks if the latest version is newer than the current version
func isUpdateAvailable(currentVersion, latestVersion string) (bool, error) {

	currentVersion = strings.TrimPrefix(currentVersion, "v")
	latestVersion = strings.TrimPrefix(latestVersion, "v")

	currentSemver, err := semver.NewVersion(currentVersion)
	if err != nil {
		return false, fmt.Errorf("parsing current version %s: %w", currentVersion, err)
	}

	latestSemver, err := semver.NewVersion(latestVersion)
	if err != nil {
		return false, fmt.Errorf("parsing latest version %s: %w", latestVersion, err)
	}

	return latestSemver.GreaterThan(currentSemver), nil
}
