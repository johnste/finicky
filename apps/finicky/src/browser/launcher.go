package browser

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"slices"

	"al.essio.dev/pkg/shellescape"
	"finicky/util"
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
	ID                string `json:"id"`
	AppName           string `json:"app_name"`
	Type              string `json:"type"`
}

func LaunchBrowser(config BrowserConfig, dryRun bool, openInBackgroundByDefault bool) error {
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

	var openInBackground bool = openInBackgroundByDefault

	if config.OpenInBackground != nil {
		openInBackground = *config.OpenInBackground
	}

	if openInBackground {
		openArgs = append(openArgs, "-g")
	}

	// Handle profile and custom args
	profileArgs, ok := resolveBrowserProfileArgs(config.Name, config.Profile)
	hasCustomArgs := len(config.Args) > 0

	// Add -n flag if profile is used (required for profile switching)
	if ok {
		openArgs = append(openArgs, "-n")
	}

	// Add --args if we have profile args or custom args
	if ok || hasCustomArgs {
		if ! slices.Contains(config.Args, "--args") {
			openArgs = append(openArgs, "--args")
		}
		// Add profile arguments first if present
		if ok {
			openArgs = append(openArgs, profileArgs...)
		}

		// Add custom args or URL
		if hasCustomArgs {
			openArgs = append(openArgs, config.Args...)
		} else {
			openArgs = append(openArgs, config.URL)
		}
	} else {
		// No special args, just add the URL
		openArgs = append(openArgs, config.URL)
	}

	cmd := exec.Command("open", openArgs...)

	// Pretty print the command with proper escaping
	prettyCmd := formatCommand(cmd.Path, cmd.Args)

	if dryRun {
		slog.Debug("Would run command (dry run)", "command", prettyCmd)
		return nil
	} else {
		slog.Debug("Run command", "command", prettyCmd)
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

func resolveBrowserProfileArgs(identifier string, profile string) ([]string, bool) {
	var browsersJson []browserInfo
	if err := json.Unmarshal(browsersJsonData, &browsersJson); err != nil {
		slog.Info("Error parsing browsers.json", "error", err)
		return nil, false
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
		return nil, false
	}

	slog.Debug("Browser found in browsers.json", "identifier", identifier, "type", matchedBrowser.Type)

	if profile != "" {
		switch matchedBrowser.Type {
		case "Chromium":
			homeDir, err := util.UserHomeDir()
			if err != nil {
				slog.Info("Error getting home directory", "error", err)
				return nil, false
			}

			localStatePath := filepath.Join(homeDir, "Library/Application Support", matchedBrowser.ConfigDirRelative, "Local State")
			if args, ok := chromiumProfileArgs(localStatePath, profile); ok {
				return args, true
			}
		case "Firefox":
			homeDir, err := util.UserHomeDir()
			if err != nil {
				slog.Info("Error getting home directory", "error", err)
				return nil, false
			}

			profilesIniPath := filepath.Join(homeDir, "Library/Application Support", matchedBrowser.ConfigDirRelative, "profiles.ini")
			if args, ok := firefoxProfileArgs(profilesIniPath, profile); ok {
				return args, true
			}
		default:
			slog.Info("Browser is not a supported browser type, skipping profile detection", "identifier", identifier)
		}
	}

	return nil, false
}

// chromiumProfileArgs resolves a configured profile to Chromium command line
// arguments, given the path to that browser's Local State.
func chromiumProfileArgs(localStatePath string, profile string) ([]string, bool) {
	profilePath, ok, denied := parseProfiles(localStatePath, profile)
	if ok {
		return []string{"--profile-directory=" + profilePath}, true
	}
	if denied && profile != "" {
		// The browser resolves the directory name itself, so hand it over
		// unresolved rather than dropping the profile. Local State maps a
		// display name to a directory and that is the file that could not be
		// read, so this only lands on the right profile when the configured
		// value is already the directory name, such as "Profile 1". Saying so
		// is better than silently opening whichever profile was last used.
		slog.Warn("Passing the profile to the browser unresolved, because the profile list could not be read", "profile", profile, "note", "this must be the profile directory name, for example \"Profile 1\", not the name shown in the browser", "suggestion", browserDataAccessHint)
		return []string{"--profile-directory=" + profile}, true
	}
	return nil, false
}

// firefoxProfileArgs resolves a configured profile to Firefox command line
// arguments, given the path to profiles.ini.
func firefoxProfileArgs(profilesIniPath string, profile string) ([]string, bool) {
	profileName, ok, denied := parseFirefoxProfiles(profilesIniPath, profile)
	if ok {
		return []string{"-P", profileName}, true
	}
	return firefoxFallbackArgs(profile, denied)
}

// browserDataAccessHint is the remedy when the operating system refuses a
// read of a browser's data directory.
//
// macOS 27 protects the application support directories of other apps, so
// reads of ~/Library/Application Support/<browser> fail with EPERM
// ("operation not permitted"). The browser itself keeps working, so the
// denial looks exactly like a profile that does not exist: nothing matches,
// the browser is launched with no profile argument, and the URL opens in
// whichever profile was last used.
//
// The grant is per browser, listed under Finicky in Files & Folders. macOS
// creates the entry, switched off, the first time it denies a read, and
// mentions it once in a "Data Access Blocked" notification. There is no
// prompt at the time and the pane has no Add button, so nothing appears
// until a denial has happened. Full Disk Access also works, but grants far
// more than this needs.
const browserDataAccessHint = "Allow Finicky to read this browser's data in System Settings > Privacy & Security > Files & Folders, then restart Finicky"

// reportBrowserDataReadError logs a failed read of a browser's data and
// reports whether the operating system refused it, which is not the same as
// the file being absent and does not mean the profile is gone.
func reportBrowserDataReadError(what string, path string, err error) bool {
	if errors.Is(err, fs.ErrPermission) {
		slog.Warn("The operating system denied access to the browser's profile data, so no profiles could be read", "file", what, "path", path, "error", err, "suggestion", browserDataAccessHint)
		return true
	}
	slog.Info("Error reading "+what, "path", path, "error", err)
	return false
}

// readFirefoxProfileNames returns the profile names in profiles.ini. The bool
// is true when the file could not be read because access was denied.
func readFirefoxProfileNames(profilesIniPath string) ([]string, bool) {
	data, err := os.ReadFile(profilesIniPath)
	if err != nil {
		return []string{}, reportBrowserDataReadError("profiles.ini", profilesIniPath, err)
	}

	names := []string{}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if name, ok := strings.CutPrefix(line, "Name="); ok {
			names = append(names, name)
		}
	}
	return names, false
}

// parseFirefoxProfiles matches a configured profile against profiles.ini. The
// second bool is true when the list could not be read because access was
// denied, so the caller can tell "no such profile" from "could not look".
func parseFirefoxProfiles(profilesIniPath string, profile string) (string, bool, bool) {
	names, denied := readFirefoxProfileNames(profilesIniPath)
	for _, name := range names {
		if name == profile {
			return name, true, false
		}
	}
	if !denied {
		slog.Warn("Could not find profile in Firefox profiles.", "Expected profile", profile, "Available profiles", strings.Join(names, ", "))
	}
	return "", false, denied
}

// firefoxFallbackArgs resolves a profile that profiles.ini could not.
//
// An absolute path names the profile directory Firefox would open anyway, and
// using it needs no access to the Firefox directory at all.
//
// Otherwise, if the list could not be read, hand the name to Firefox
// unresolved. Firefox resolves "-P" itself through GetProfileByName against
// the profiles.ini it reads directly, and Firefox is not the process being
// denied, so the lookup succeeds. A name Firefox cannot find opens the
// profile manager, which is a visible failure the user can act on, rather
// than the silent one this replaces. When the list was readable and the name
// still missed, the name is simply wrong and is left unresolved.
func firefoxFallbackArgs(profile string, denied bool) ([]string, bool) {
	if profile == "" {
		return nil, false
	}

	if filepath.IsAbs(profile) {
		info, err := os.Stat(profile)
		switch {
		case err == nil && !info.IsDir():
			slog.Warn("Firefox profile path is not a directory", "path", profile)
			return nil, false
		case errors.Is(err, fs.ErrNotExist):
			// Firefox creates an empty profile for a directory that is not
			// there rather than failing, which loses a session silently.
			slog.Warn("Firefox profile path does not exist", "path", profile, "note", "Firefox would create an empty profile there, so it is not used")
			return nil, false
		case err != nil:
			// Under a denial the directory is almost certainly there, and
			// Firefox is not the one being denied.
			slog.Warn("Using Firefox profile path that could not be checked", "path", profile, "error", err, "suggestion", browserDataAccessHint)
		default:
			slog.Info("Using Firefox profile path directly", "path", profile)
		}
		return []string{"--profile", profile}, true
	}

	if denied {
		slog.Warn("Passing the Firefox profile name to Firefox unresolved, because the profile list could not be read", "name", profile, "suggestion", browserDataAccessHint)
		return []string{"-P", profile}, true
	}
	return nil, false
}

// chromiumInfoCache reads the profile list from Local State. The last bool is
// true when the file could not be read because access was denied.
func chromiumInfoCache(localStatePath string) (map[string]interface{}, bool, bool) {
	data, err := os.ReadFile(localStatePath)
	if err != nil {
		return nil, false, reportBrowserDataReadError("Local State file", localStatePath, err)
	}

	var localState map[string]interface{}
	if err := json.Unmarshal(data, &localState); err != nil {
		slog.Info("Error parsing Local State JSON", "error", err)
		return nil, false, false
	}

	profiles, ok := localState["profile"].(map[string]interface{})
	if !ok {
		slog.Info("Could not find profile section in Local State")
		return nil, false, false
	}

	infoCache, ok := profiles["info_cache"].(map[string]interface{})
	if !ok {
		slog.Info("Could not find info_cache in profile section")
		return nil, false, false
	}

	return infoCache, true, false
}

func getAllChromiumProfiles(localStatePath string) []string {
	cache, ok, _ := chromiumInfoCache(localStatePath)
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

// parseProfiles matches a configured profile name against Local State. The
// last bool is true when the list could not be read because access was
// denied, so the caller can tell "no such profile" from "could not look".
func parseProfiles(localStatePath string, profile string) (string, bool, bool) {
	infoCache, ok, denied := chromiumInfoCache(localStatePath)
	if !ok {
		return "", false, denied
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
			return profilePath, true, false
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
			return profilePath, true, false
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

	return "", false, false
}

// GetProfilesForBrowser returns available profile names for a given browser app name or bundle ID.
// Returns empty slice if browser not in browsers.json, not supported, or profile files are unreadable.
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

	homeDir, err := util.UserHomeDir()
	if err != nil {
		slog.Info("Error getting home directory", "error", err)
		return []string{}
	}

	switch matchedBrowser.Type {
	case "Chromium":
		localStatePath := filepath.Join(homeDir, "Library/Application Support", matchedBrowser.ConfigDirRelative, "Local State")
		return getAllChromiumProfiles(localStatePath)
	case "Firefox":
		profilesIniPath := filepath.Join(homeDir, "Library/Application Support", matchedBrowser.ConfigDirRelative, "profiles.ini")
		names, _ := readFirefoxProfileNames(profilesIniPath)
		return names
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
