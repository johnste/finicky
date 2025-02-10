package version

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dop251/goja"
)

type GithubRelease struct {
	TagName    string `json:"tag_name"`
	Name       string `json:"name"`
	Prerelease bool   `json:"prerelease"`
	Draft      bool   `json:"draft"`
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
		log.Printf("Error getting user home directory: %v", err)
		return ""
	}
	cacheDir := filepath.Join(homeDir, "Library", "Caches", "Finicky")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		log.Printf("Error creating cache directory: %v", err)
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
		log.Printf("Error saving last update check time: %v", err)
	}
}

func GetCurrentVersion() string {
	// Get the bundle path
	bundlePath := os.Getenv("BUNDLE_PATH")
	if bundlePath == "" {
		execPath, err := os.Executable()
		if err != nil {
			log.Printf("Error getting executable path: %v", err)
			return ""
		}
		// Assuming the standard macOS app bundle structure
		bundlePath = filepath.Join(filepath.Dir(execPath), "..", "Info.plist")
	}

	// Read and parse Info.plist
	cmd := exec.Command("defaults", "read", bundlePath, "CFBundleShortVersionString")
	output, err := cmd.Output()
	if err != nil {
		log.Printf("Error reading version from Info.plist: %v", err)
		return ""
	}

	return strings.TrimSpace(string(output))
}

func CheckForUpdatesAsync() {
	log.Println("Checking update schedule...")

	lastCheck := getLastUpdateCheck()
	if !lastCheck.IsZero() {
		timeSinceLastCheck := time.Since(lastCheck)
		if timeSinceLastCheck < 24*time.Hour {
			log.Printf("Skipping update check - last checked %v ago", timeSinceLastCheck.Round(time.Minute))
			return
		}
	}

	log.Println("Checking for updates...")
	currentVersion := GetCurrentVersion()
	if currentVersion == "" {
		log.Println("Could not determine current version")
		return
	}
	log.Printf("Current version: %s", currentVersion)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", "https://api.github.com/repos/johnste/finicky/releases", nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}

	// Set User-Agent header
	req.Header.Set("User-Agent", fmt.Sprintf("finicky/%s", currentVersion))

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error making request: %v", err)
		return
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response: %v", err)
		return
	}

	// Parse releases
	var releases []GithubRelease
	if err := json.Unmarshal(body, &releases); err != nil {
		log.Printf("Error parsing releases: %v %s", err, string(body))
		return
	}

	// Filter out prereleases and drafts
	var stableReleases []GithubRelease
	for _, release := range releases {
		if !release.Prerelease && !release.Draft {
			stableReleases = append(stableReleases, release)
		}
	}

	if len(stableReleases) == 0 {
		log.Println("No stable releases found")
		return
	}

	// Get latest version (first in the list)
	latestRelease := stableReleases[0]
	latestVersion := strings.TrimPrefix(latestRelease.TagName, "v")

	log.Printf("Latest version available: %s", latestVersion)

	// Compare versions using semantic versioning
	if compareSemver(latestVersion, currentVersion) > 0 {
		log.Printf("New version %s is available! (current: %s)", latestVersion, currentVersion)
		// TODO: Implement update notification mechanism
	} else {
		log.Println("You are running the latest version")
	}

	// Update the last check time
	setLastUpdateCheck(time.Now())
}

// CheckForUpdatesFromConfig checks if updates should be performed based on VM configuration
func CheckForUpdatesFromConfig(vm *goja.Runtime) error {
	// Check checkForUpdates option
	checkForUpdates, err := vm.RunString("finickyConfigAPI.getOption('checkForUpdates', mergedConfig)")
	if err != nil {
		return fmt.Errorf("failed to get checkForUpdates option: %v", err)
	}

	if checkForUpdates.ToBoolean() {
		CheckForUpdatesAsync()
	}
	return nil
}