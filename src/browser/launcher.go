package browser

import (
	_ "embed"
	"encoding/json"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed browsers.json
var browsersJsonData []byte

type BrowserResult struct {
	Browser BrowserConfig `json:"browser"`
	Error string `json:"error"`
}

type BrowserConfig struct {
	Name             string   `json:"name"`
	AppType          string   `json:"appType"`
	OpenInBackground bool     `json:"openInBackground"`
	Profile          string   `json:"profile"`
	Args             []string `json:"args"`
	URL              string   `json:"url"`
}

type browserInfo struct {
	ConfigDirRelative string `json:"config_dir_relative"`
	ID               string `json:"id"`
	Type             string `json:"type"`
}

func LaunchBrowser(config BrowserConfig) error {
	appName := config.Name
	useBundleId := config.AppType == "bundleId"

	if config.AppType == "none" {
		slog.Info("AppType is 'none', not launching any browser")
		return nil
	}

	slog.Info("Starting browser", "name", config.Name, "url", config.URL)

	openArgs := []string{"-a", appName}
	if config.AppType == "bundleId" {
		openArgs = []string{"-b", appName}
	}

	if config.OpenInBackground {
		openArgs = append(openArgs, "-g")
	}

	if len(config.Args) == 0 {
		if useBundleId {
			profileArgument, ok := resolveBrowserProfileArgument(appName, config.Profile)
			if ok {
				// FIXME: This is a hack to get the profile argument to work – this won't work for Firefox
				openArgs = append(openArgs, "-n")
				openArgs = append(openArgs, "--args")
				openArgs = append(openArgs, profileArgument)
			}
		}

		openArgs = append(openArgs, config.URL)
	} else {
		openArgs = append(openArgs, config.Args...)
	}

	cmd := exec.Command("open", openArgs...)

	if os.Getenv("FINICKY_DRY_RUN") == "true" {
		slog.Info("Would run command (dry run)", "command", cmd.String())
		return nil
	} else {
		slog.Info("Run command", "command", cmd.String())
	}

	return cmd.Start()
}

func resolveBrowserProfileArgument(bundleId string, profile string) (string, bool) {
	var browsersJson []browserInfo
	if err := json.Unmarshal(browsersJsonData, &browsersJson); err != nil {
		slog.Info("Error parsing browsers.json", "error", err)
		return "", false
	}

	// Try to find matching browser by bundle ID
	var matchedBrowser *browserInfo
	for _, browser := range browsersJson {
		if browser.ID == bundleId {
			matchedBrowser = &browser
			break
		}
	}

	if matchedBrowser == nil {
		return "", false
	}

	slog.Info("Browser found in browsers.json", "bundleId", bundleId, "type", matchedBrowser.Type)

	if profile != ""  {
		switch matchedBrowser.Type {
		case "Chromium":
			homeDir, err := os.UserHomeDir()
			if err != nil {
				slog.Info("Error getting home directory", "error", err)
				return "", false
			}

			localStatePath := filepath.Join(homeDir, "Library/Application Support", matchedBrowser.ConfigDirRelative, "Local State")
			profilePath, ok := parseProfiles(localStatePath, profile)
			if ok {
				return "--profile-directory=" + profilePath, true
			}
		default:
			slog.Info("Browser is not a Chromium browser, skipping profile detection", "bundleId", bundleId)
		}
	}

	return "", true
}

func parseProfiles(localStatePath string, profile string) (string, bool) {
	data, err := os.ReadFile(localStatePath)
	if err != nil {
		slog.Info("Error reading Local State file", "path", localStatePath, "error", err)
		return "", false
	}

	var localState map[string]interface{}
	if err := json.Unmarshal(data, &localState); err != nil {
		slog.Info("Error parsing Local State JSON", "error", err)
		return "", false
	}

	profiles, ok := localState["profile"].(map[string]interface{})
	if !ok {
		slog.Info("Could not find profile section in Local State")
		return "", false
	}

	infoCache, ok := profiles["info_cache"].(map[string]interface{})
	if !ok {
		slog.Info("Could not find info_cache in profile section")
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
	slog.Info("Could not find profile in browser profiles, trying to find by profile path", "profile", profile)
	for profilePath, info := range infoCache {
		if profilePath == profile {
			// Try to get the profile name of the profile we want the user to use instead
			if profileInfo, ok := info.(map[string]interface{}); ok {
				if name, ok := profileInfo["name"].(string); ok {
					slog.Info("Found profile using profile path", "path", profilePath, "name", name, "suggestion", "Please use the profile name instead")
				}
			}
			return profilePath, true
		}
	}

	slog.Info("Could not find profile in browser profiles", "profile", profile)
	slog.Info("Available profiles:")
	for _, info := range infoCache {
		profileInfo, ok := info.(map[string]interface{})
		if !ok {
			continue
		}

		name, ok := profileInfo["name"].(string)
		if !ok {
			continue
		}

		slog.Info("- Profile", "name", name)
	}

	return "", false
}
