package browser

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	AppName          string `json:"app_name"`
	Type             string `json:"type"`
}

func LaunchBrowser(config BrowserConfig, dryRun bool) error {
	if config.AppType == "none" {
		slog.Info("AppType is 'none', not launching any browser")
		return nil
	}

	slog.Info("Starting browser", "name", config.Name, "url", config.URL)

	var openArgs []string

	if config.AppType == "bundleId" {
		openArgs = []string{"-b", config.Name}
	} else {
		openArgs = []string{"-a", config.Name}
	}

	if config.OpenInBackground {
		openArgs = append(openArgs, "-g")
	}

	if len(config.Args) == 0 {

		profileArgument, ok := resolveBrowserProfileArgument(config.Name, config.Profile)
		if ok {
			// FIXME: This is a hack to get the profile argument to work – this won't work for Firefox
			openArgs = append(openArgs, "-n")
			openArgs = append(openArgs, "--args")
			openArgs = append(openArgs, profileArgument)
		}

		openArgs = append(openArgs, config.URL)
	} else {
		openArgs = append(openArgs, config.Args...)
	}

	cmd := exec.Command("open", openArgs...)

	if dryRun {
		slog.Debug("Would run command (dry run)", "command", cmd.String())
		return nil
	} else {
		slog.Debug("Run command", "command", cmd.String())
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	stderrBytes, err := io.ReadAll(stderr)
	if err != nil {
		return fmt.Errorf("error reading stderr: %v", err)
	}

	stdoutBytes, err := io.ReadAll(stdout)
	if err != nil {
		return fmt.Errorf("error reading stdout: %v", err)
	}

	cmdErr := cmd.Wait()

	if len(stderrBytes) > 0 {
		slog.Error("Command returned error", "error", string(stderrBytes))
	}
	if len(stdoutBytes) > 0 {
		slog.Debug("Command returned output", "output", string(stdoutBytes))
	}

	if cmdErr != nil {
		return fmt.Errorf("command failed: %v", cmdErr)
	}

	return nil
}

func resolveBrowserProfileArgument(identifier string, profile string) (string, bool) {
	var browsersJson []browserInfo
	if err := json.Unmarshal(browsersJsonData, &browsersJson); err != nil {
		slog.Info("Error parsing browsers.json", "error", err)
		return "", false
	}

	// Try to find matching browser by bundle ID
	var matchedBrowser *browserInfo
	for _, browser := range browsersJson {
		if browser.ID == identifier || browser.AppName == identifier {
			matchedBrowser = &browser
			break
		}
	}

	if matchedBrowser == nil {
		return "", false
	}

	slog.Info("Browser found in browsers.json", "identifier", identifier, "type", matchedBrowser.Type)

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
				return "--profile-directory=\"" + profilePath + "\"", true
			}
		default:
			slog.Info("Browser is not a Chromium browser, skipping profile detection", "identifier", identifier)
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
