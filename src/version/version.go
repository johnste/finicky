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
	"strconv"
	"strings"
	"time"

	"github.com/dop251/goja"
)

const updateCheckInterval = 24 * time.Hour

type GithubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
}

var (
	// These will be set via ldflags during build
	commitHash = "dev"
	buildDate = "unknown"
)

// GetBuildInfo returns the commit hash and build date
func GetBuildInfo() (string, string) {
	return commitHash, buildDate
}

// compareSemver compares two semantic version strings.
// Returns:
//   -1 if v1 < v2
//    0 if v1 == v2
//    1 if v1 > v2
func compareSemver(v1, v2 string) int {
	// Split version strings into components
	v1Parts := strings.Split(strings.TrimPrefix(v1, "v"), ".")
	v2Parts := strings.Split(strings.TrimPrefix(v2, "v"), ".")

	// Compare each component
	for i := 0; i < len(v1Parts) && i < len(v2Parts); i++ {
		n1, err1 := strconv.Atoi(v1Parts[i])
		n2, err2 := strconv.Atoi(v2Parts[i])

		// If either part isn't a valid number, do string comparison
		if err1 != nil || err2 != nil {
			if v1Parts[i] < v2Parts[i] {
				return -1
			}
			if v1Parts[i] > v2Parts[i] {
				return 1
			}
			continue
		}

		if n1 < n2 {
			return -1
		}
		if n1 > n2 {
			return 1
		}
	}

	// If all components so far are equal, longer version is considered greater
	if len(v1Parts) < len(v2Parts) {
		return -1
	}
	if len(v1Parts) > len(v2Parts) {
		return 1
	}

	return 0
}

func getCacheDir() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		slog.Error("Error getting user home directory", "error", err)
		return ""
	}
	cacheDir := filepath.Join(homeDir, "Library", "Caches", "Finicky")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		slog.Error("Error creating cache directory", "error", err)
		return ""
	}
	return cacheDir
}

func getLastUpdateCheck() time.Time {
	cacheDir := getCacheDir()
	if cacheDir == "" {
		return time.Time{}
	}

	cacheFile := filepath.Join(cacheDir, "last_update_check")
	data, err := os.ReadFile(cacheFile)
	if err != nil {
		// If file doesn't exist or can't be read, return zero time
		return time.Time{}
	}

	timestamp, err := strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
	if err != nil {
		return time.Time{}
	}

	return time.Unix(timestamp, 0)
}

func setLastUpdateCheck(t time.Time) {
	cacheDir := getCacheDir()
	if cacheDir == "" {
		return
	}

	cacheFile := filepath.Join(cacheDir, "last_update_check")
	timestamp := fmt.Sprintf("%d", t.Unix())
	if err := os.WriteFile(cacheFile, []byte(timestamp), 0644); err != nil {
		slog.Error("Error saving last update check time", "error", err)
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

func checkForUpdates() {
	slog.Debug("Checking update schedule...")

	lastCheck := getLastUpdateCheck()
	if !lastCheck.IsZero() {
		timeSinceLastCheck := time.Since(lastCheck)
		if timeSinceLastCheck < updateCheckInterval {
			slog.Debug("Skipping update check - last checked", "duration", fmt.Sprintf("%d minutes", int(timeSinceLastCheck.Minutes())))
			return
		}
	}

	slog.Info("Checking for updates...")
	currentVersion := GetCurrentVersion()
	if currentVersion == "" {
		slog.Info("Could not determine current version")
		return
	}
	slog.Info("Current version", "version", currentVersion)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", "https://api.github.com/repos/johnste/finicky/releases", nil)
	if err != nil {
		slog.Error("Error creating request", "error", err)
		return
	}

	// Set User-Agent header
	req.Header.Set("User-Agent", fmt.Sprintf("finicky/%s", currentVersion))

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		slog.Error("Error making request", "error", err)
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		slog.Error("Error reading response", "error", err)
		return
	}

	// Parse releases
	var releases []GithubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		slog.Error("Error parsing releases", "error", err, "response", string(body))
		return
	}

	// Check if current version is a prerelease (contains alpha/beta)
	currentIsPrerelease := strings.Contains(strings.ToLower(currentVersion), "alpha") ||
		strings.Contains(strings.ToLower(currentVersion), "beta")

	if currentIsPrerelease {
		slog.Info("Currently using a prerelease version, checking for pending prerelease releases", "version", currentVersion)
	}

	// Filter out prereleases and drafts
	var pendingReleases []GithubRelease
	for _, release := range releases {
		if !release.Prerelease && !release.Draft || currentIsPrerelease && release.Prerelease {
			pendingReleases = append(pendingReleases, release)
		}
	}

	if len(pendingReleases) == 0 {
		slog.Info("No stable releases found")
		return
	}

	// Get latest version (first in the list)
	latestRelease := pendingReleases[0]
	latestVersion := strings.TrimPrefix(latestRelease.TagName, "v")

	slog.Debug("Latest version available", "version", latestVersion)

	// Compare versions using semantic versioning
	if compareSemver(latestVersion, currentVersion) > 0 {
		slog.Info("New version is available", "latestVersion", latestVersion, "currentVersion", currentVersion)
		// TODO: Implement update notification mechanism
	} else {
		slog.Info("You are running the latest version")
	}

	// Update the last check time
	setLastUpdateCheck(time.Now())
}

// CheckForUpdatesFromConfig checks if updates should be performed based on VM configuration
func CheckForUpdatesFromConfig(vm *goja.Runtime) error {

	if vm == nil {
		// Check for updates if we don't have a VM
		checkForUpdates()
		return nil
	}

	// Check checkForUpdates option
	shouldCheckForUpdates, err := vm.RunString("finickyConfigAPI.getOption('checkForUpdates', finalConfig)")
	if err != nil {
		return fmt.Errorf("failed to get checkForUpdates option: %v", err)
	}

	if shouldCheckForUpdates.ToBoolean() {
		checkForUpdates()
	} else {
		slog.Debug("Skipping update check")
	}
	return nil
}