package browser

import (
	_ "embed"
	"encoding/json"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"al.essio.dev/pkg/shellescape"
)

//go:embed browsers.json
var browsersJsonData []byte

type BrowserResult struct {
	Browser BrowserConfig `json:"browser"`
	Error   string        `json:"error"`
}

type BrowserConfig struct {
	Name             string   `json:"name"`
	AppType          string   `json:"appType"`
	OpenInBackground *bool    `json:"openInBackground"`
	Profile          string   `json:"profile"`
	Args             []string `json:"args"`
	URL              string   `json:"url"`
}

type browserInfo struct {
	ConfigDirRelative string   `json:"config_dir_relative"`
	ID                string   `json:"id"`
	AppName           string   `json:"app_name"`
	Type              string   `json:"type"`
	LinuxConfigDirs   []string `json:"linux_config_dirs"`
	LinuxDesktopIDs   []string `json:"linux_desktop_ids"`
}

func LaunchBrowser(config BrowserConfig, dryRun bool, openInBackgroundByDefault bool) error {
	if config.AppType == "none" {
		slog.Info("AppType is 'none', not launching any browser")
		return nil
	}

	slog.Info("Starting browser", "name", config.Name, "url", config.URL)

	var openInBackground bool = openInBackgroundByDefault

	if config.OpenInBackground != nil {
		openInBackground = *config.OpenInBackground
	}

	profileArgs, hasProfile := resolveBrowserProfileArgs(config.Name, config.Profile)

	cmd, err := browserCommand(config, openInBackground, profileArgs, hasProfile)
	if err != nil {
		return err
	}

	// Pretty print the command with proper escaping
	prettyCmd := formatCommand(cmd.Path, cmd.Args)

	if dryRun {
		slog.Debug("Would run command (dry run)", "command", prettyCmd)
		return nil
	} else {
		slog.Debug("Run command", "command", prettyCmd)
	}

	return runBrowserCommand(cmd)
}

// lookupBrowserInfo finds the browsers.json entry for a browser app name or
// bundle ID (or, on Linux, its desktop entry).
func lookupBrowserInfo(identifier string) (*browserInfo, bool) {
	var browsersJson []browserInfo
	if err := json.Unmarshal(browsersJsonData, &browsersJson); err != nil {
		slog.Info("Error parsing browsers.json", "error", err)
		return nil, false
	}

	for i := range browsersJson {
		if browsersJson[i].ID == identifier || browsersJson[i].AppName == identifier {
			return &browsersJson[i], true
		}
	}
	for i := range browsersJson {
		if platformBrowserMatches(browsersJson[i], identifier) {
			return &browsersJson[i], true
		}
	}
	return nil, false
}

func resolveBrowserProfileArgs(identifier string, profile string) ([]string, bool) {
	matchedBrowser, ok := lookupBrowserInfo(identifier)
	if !ok {
		return nil, false
	}

	slog.Debug("Browser found in browsers.json", "identifier", identifier, "type", matchedBrowser.Type)

	if profile != "" {
		root, ok := profileRoot(*matchedBrowser)
		if !ok {
			slog.Info("No profile directory known for browser on this platform", "identifier", identifier)
			return nil, false
		}
		switch matchedBrowser.Type {
		case "Chromium":
			profilePath, ok := parseProfiles(filepath.Join(root, "Local State"), profile)
			if ok {
				return []string{"--profile-directory=" + profilePath}, true
			}
		case "Firefox":
			profileName, ok := parseFirefoxProfiles(filepath.Join(root, "profiles.ini"), profile)
			if ok {
				return []string{"-P", profileName}, true
			}
		default:
			slog.Info("Browser is not a supported browser type, skipping profile detection", "identifier", identifier)
		}
	}

	return nil, false
}

func readFirefoxProfileNames(profilesIniPath string) []string {
	data, err := os.ReadFile(profilesIniPath)
	if err != nil {
		slog.Info("Error reading profiles.ini", "path", profilesIniPath, "error", err)
		return []string{}
	}

	names := []string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if name, ok := strings.CutPrefix(line, "Name="); ok {
			names = append(names, name)
		}
	}
	return names
}

func parseFirefoxProfiles(profilesIniPath string, profile string) (string, bool) {
	names := readFirefoxProfileNames(profilesIniPath)
	for _, name := range names {
		if name == profile {
			return name, true
		}
	}
	slog.Warn("Could not find profile in Firefox profiles.", "Expected profile", profile, "Available profiles", strings.Join(names, ", "))
	return "", false
}

func chromiumInfoCache(localStatePath string) (map[string]interface{}, bool) {
	data, err := os.ReadFile(localStatePath)
	if err != nil {
		slog.Info("Error reading Local State file", "path", localStatePath, "error", err)
		return nil, false
	}

	var localState map[string]interface{}
	if err := json.Unmarshal(data, &localState); err != nil {
		slog.Info("Error parsing Local State JSON", "error", err)
		return nil, false
	}

	profiles, ok := localState["profile"].(map[string]interface{})
	if !ok {
		slog.Info("Could not find profile section in Local State")
		return nil, false
	}

	infoCache, ok := profiles["info_cache"].(map[string]interface{})
	if !ok {
		slog.Info("Could not find info_cache in profile section")
		return nil, false
	}

	return infoCache, true
}

func getAllChromiumProfiles(localStatePath string) []string {
	cache, ok := chromiumInfoCache(localStatePath)
	if !ok {
		return []string{}
	}

	var names []string
	for _, info := range cache {
		profileInfo, ok := info.(map[string]interface{})
		if !ok {
			continue
		}
		name, ok := profileInfo["name"].(string)
		if !ok {
			continue
		}
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func parseProfiles(localStatePath string, profile string) (string, bool) {
	infoCache, ok := chromiumInfoCache(localStatePath)
	if !ok {
		return "", false
	}

	// Look for the specified profile
	for profilePath, info := range infoCache {
		profileInfo, ok := info.(map[string]interface{})
		if !ok {
			continue
		}

		name, ok := profileInfo["name"].(string)
		if !ok {
			continue
		}

		if name == profile {
			slog.Info("Found profile by name", "name", name, "path", profilePath)
			return profilePath, true
		}
	}

	// If we didn't find the profile, try to find it by profile folder name
	slog.Debug("Could not find profile in browser profiles, trying to find by profile path", "profile", profile)
	for profilePath, info := range infoCache {
		if profilePath == profile {
			// Try to get the profile name of the profile we want the user to use instead
			if profileInfo, ok := info.(map[string]interface{}); ok {
				if name, ok := profileInfo["name"].(string); ok {
					slog.Warn("Found profile using profile path", "path", profilePath, "name", name, "suggestion", "Please use the profile name instead")
				}
			}
			return profilePath, true
		}
	}

	var profileNames []string
	for _, info := range infoCache {
		profileInfo, ok := info.(map[string]interface{})
		if !ok {
			continue
		}

		name, ok := profileInfo["name"].(string)
		if !ok {
			continue
		}

		profileNames = append(profileNames, name)
	}
	slog.Warn("Could not find profile in browser profiles.", "Expected profile", profile, "Available profiles", strings.Join(profileNames, ", "))

	return "", false
}

// GetProfilesForBrowser returns available profile names for a given browser app name or bundle ID.
// Returns empty slice if browser not in browsers.json, not supported, or profile files are unreadable.
func GetProfilesForBrowser(identifier string) []string {
	matchedBrowser, ok := lookupBrowserInfo(identifier)
	if !ok {
		return []string{}
	}

	root, ok := profileRoot(*matchedBrowser)
	if !ok {
		return []string{}
	}

	switch matchedBrowser.Type {
	case "Chromium":
		return getAllChromiumProfiles(filepath.Join(root, "Local State"))
	case "Firefox":
		return readFirefoxProfileNames(filepath.Join(root, "profiles.ini"))
	default:
		return []string{}
	}
}

// formatCommand returns a properly shell-escaped string representation of the command
func formatCommand(path string, args []string) string {
	if len(args) == 0 {
		return shellescape.Quote(path)
	}

	quotedArgs := make([]string, len(args))
	for i, arg := range args {
		quotedArgs[i] = shellescape.Quote(arg)
	}

	return strings.Join(quotedArgs, " ")
}
