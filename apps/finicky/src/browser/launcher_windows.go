//go:build windows

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
	ConfigDirRelative string `json:"config_dir_relative"`
	ConfigDirWindows  string `json:"config_dir_windows"`
	ID                string `json:"id"`
	AppName           string `json:"app_name"`
	Type              string `json:"type"`
}

func (b browserInfo) windowsConfigDir() string {
	if b.ConfigDirWindows != "" {
		return b.ConfigDirWindows
	}
	return b.ConfigDirRelative
}

func LaunchBrowser(config BrowserConfig, dryRun bool, openInBackgroundByDefault bool) error {
	if config.AppType == "none" {
		slog.Info("AppType is 'none', not launching any browser")
		return nil
	}

	slog.Info("Starting browser", "name", config.Name, "url", config.URL)

	profileArgs, hasProfile := resolveBrowserProfileArgs(config.Name, config.Profile)
	hasCustomArgs := len(config.Args) > 0

	var cmdArgs []string

	if hasProfile || hasCustomArgs {
		if hasProfile {
			cmdArgs = append(cmdArgs, profileArgs...)
		}
		if hasCustomArgs {
			cmdArgs = append(cmdArgs, config.Args...)
		}
		if !slices.Contains(config.Args, config.URL) {
			cmdArgs = append(cmdArgs, config.URL)
		}
	} else {
		cmdArgs = append(cmdArgs, config.URL)
	}

	// Find the browser executable
	exePath := findBrowserExe(config.Name)
	if exePath == "" {
		// Fallback: use cmd /c start to let Windows find the browser
		allArgs := append([]string{"/c", "start", "", config.URL}, cmdArgs...)
		cmd := exec.Command("cmd", allArgs...)
		prettyCmd := formatCommand(cmd.Path, cmd.Args)
		if dryRun {
			slog.Debug("Would run command (dry run)", "command", prettyCmd)
			return nil
		}
		slog.Debug("Run command (fallback)", "command", prettyCmd)
		return cmd.Run()
	}

	cmd := exec.Command(exePath, cmdArgs...)
	prettyCmd := formatCommand(cmd.Path, cmd.Args)

	if dryRun {
		slog.Debug("Would run command (dry run)", "command", prettyCmd)
		return nil
	}

	slog.Debug("Run command", "command", prettyCmd)

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

	stderrBytes, _ := io.ReadAll(stderr)
	stdoutBytes, _ := io.ReadAll(stdout)
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

func findBrowserExe(name string) string {
	commonPaths := []string{
		os.Getenv("PROGRAMFILES"),
		os.Getenv("PROGRAMFILES(X86)"),
		os.Getenv("LOCALAPPDATA"),
	}

	knownExes := map[string][]string{
		"Google Chrome":    {"Google\\Chrome\\Application\\chrome.exe"},
		"Chrome":           {"Google\\Chrome\\Application\\chrome.exe"},
		"Microsoft Edge":   {"Microsoft\\Edge\\Application\\msedge.exe"},
		"Edge":             {"Microsoft\\Edge\\Application\\msedge.exe"},
		"Firefox":          {"Mozilla Firefox\\firefox.exe"},
		"Mozilla Firefox":  {"Mozilla Firefox\\firefox.exe"},
		"Brave Browser":    {"BraveSoftware\\Brave-Browser\\Application\\brave.exe"},
		"Brave":            {"BraveSoftware\\Brave-Browser\\Application\\brave.exe"},
		"Vivaldi":          {"Vivaldi\\Application\\vivaldi.exe"},
		"Opera":            {"Opera\\opera.exe"},
		"Arc":              {"Arc\\Arc.exe"},
	}

	if paths, ok := knownExes[name]; ok {
		for _, base := range commonPaths {
			if base == "" {
				continue
			}
			for _, rel := range paths {
				full := filepath.Join(base, rel)
				if _, err := os.Stat(full); err == nil {
					return full
				}
			}
		}
	}

	return ""
}

func resolveBrowserProfileArgs(identifier string, profile string) ([]string, bool) {
	var browsersJson []browserInfo
	if err := json.Unmarshal(browsersJsonData, &browsersJson); err != nil {
		slog.Info("Error parsing browsers.json", "error", err)
		return nil, false
	}

	var matchedBrowser *browserInfo
	for _, browser := range browsersJson {
		if browser.ID == identifier || browser.AppName == identifier {
			matchedBrowser = &browser
			break
		}
	}

	if matchedBrowser == nil {
		return nil, false
	}

	slog.Debug("Browser found in browsers.json", "identifier", identifier, "type", matchedBrowser.Type)

	if profile != "" {
		switch matchedBrowser.Type {
		case "Chromium":
			localAppData := os.Getenv("LOCALAPPDATA")
			if localAppData == "" {
				slog.Info("LOCALAPPDATA not set")
				return nil, false
			}
			localStatePath := filepath.Join(localAppData, matchedBrowser.windowsConfigDir(), "Local State")
			profilePath, ok := parseProfiles(localStatePath, profile)
			if ok {
				return []string{"--profile-directory=" + profilePath}, true
			}
		case "Firefox":
			appData := os.Getenv("APPDATA")
			if appData == "" {
				slog.Info("APPDATA not set")
				return nil, false
			}
			profilesIniPath := filepath.Join(appData, matchedBrowser.windowsConfigDir(), "profiles.ini")
			profileName, ok := parseFirefoxProfiles(profilesIniPath, profile)
			if ok {
				return []string{"-P", profileName}, true
			}
		default:
			slog.Info("Browser is not a supported browser type", "identifier", identifier)
		}
	}

	return nil, false
}

func GetProfilesForBrowser(identifier string) []string {
	var browsersJson []browserInfo
	if err := json.Unmarshal(browsersJsonData, &browsersJson); err != nil {
		slog.Info("Error parsing browsers.json", "error", err)
		return []string{}
	}

	var matchedBrowser *browserInfo
	for i := range browsersJson {
		if browsersJson[i].ID == identifier || browsersJson[i].AppName == identifier {
			matchedBrowser = &browsersJson[i]
			break
		}
	}

	if matchedBrowser == nil {
		return []string{}
	}

	switch matchedBrowser.Type {
	case "Chromium":
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData == "" {
			return []string{}
		}
		localStatePath := filepath.Join(localAppData, matchedBrowser.windowsConfigDir(), "Local State")
		return getAllChromiumProfiles(localStatePath)
	case "Firefox":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return []string{}
		}
		profilesIniPath := filepath.Join(appData, matchedBrowser.windowsConfigDir(), "profiles.ini")
		return readFirefoxProfileNames(profilesIniPath)
	default:
		return []string{}
	}
}

// --- Shared profile parsing (duplicated from launcher.go, can be extracted later) ---

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
	slog.Debug("Could not find profile in browser profiles, trying to find by profile path", "profile", profile)
	for profilePath, info := range infoCache {
		if profilePath == profile {
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

