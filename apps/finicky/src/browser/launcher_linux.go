package browser

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"

	"finicky/util"
)

// browserCommand builds the command that launches the browser from its
// desktop entry's Exec line, falling back to an executable of that name.
// Profile arguments come first, then the custom args or the URL, matching
// the macOS behavior where custom args replace the URL.
func browserCommand(config BrowserConfig, openInBackground bool, profileArgs []string, hasProfile bool) (*exec.Cmd, error) {
	if openInBackground {
		slog.Debug("Opening in the background is not supported on Linux")
	}

	argv, err := browserArgv(config.Name)
	if err != nil {
		return nil, err
	}

	if hasProfile {
		argv = append(argv, profileArgs...)
	}
	if len(config.Args) > 0 {
		// "--args" separates open's own arguments from the app's on macOS;
		// there's no such separation here.
		for _, arg := range config.Args {
			if arg != "--args" {
				argv = append(argv, arg)
			}
		}
	} else {
		argv = append(argv, config.URL)
	}

	return exec.Command(argv[0], argv[1:]...), nil
}

// browserArgv returns the command line that starts the browser, without a URL.
func browserArgv(name string) ([]string, error) {
	if entry, ok := findBrowserEntry(loadDesktopEntries(applicationDirs()), name); ok {
		slog.Debug("Found desktop entry for browser", "name", name, "id", entry.ID)
		return parseExec(entry.Exec, "")
	}
	if path, err := exec.LookPath(name); err == nil {
		return []string{path}, nil
	}
	return nil, fmt.Errorf("browser %q not found: no desktop entry or executable with that name", name)
}

// findBrowserEntry finds the desktop entry for a browser name, desktop ID or
// executable, or for a browsers.json bundle ID or app name (so configs
// written for macOS, e.g. "Google Chrome" or "org.mozilla.firefox", work).
func findBrowserEntry(entries []desktopEntry, name string) (desktopEntry, bool) {
	if entry, ok := findDesktopEntry(entries, name); ok {
		return entry, true
	}
	info, ok := lookupBrowserInfo(name)
	if !ok {
		return desktopEntry{}, false
	}
	for _, id := range info.LinuxDesktopIDs {
		if entry, ok := findDesktopEntry(entries, id); ok {
			return entry, true
		}
	}
	return desktopEntry{}, false
}

// runBrowserCommand starts the browser without waiting for it, since unlike
// macOS's `open` the command is the browser process itself. It runs in its
// own session so it outlives Finicky and isn't killed with Finicky's
// terminal or process group.
func runBrowserCommand(cmd *exec.Cmd) error {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start browser: %w", err)
	}
	go func() {
		if err := cmd.Wait(); err != nil {
			slog.Debug("Browser process exited", "error", err)
		}
	}()
	return nil
}

// profileRoot returns the directory holding the browser's profiles
// ("Local State" for Chromium browsers, profiles.ini for Firefox): the first
// of the browser's known Linux config directories that exists.
func profileRoot(info browserInfo) (string, bool) {
	homeDir, err := util.UserHomeDir()
	if err != nil {
		slog.Info("Error getting home directory", "error", err)
		return "", false
	}
	for _, dir := range info.LinuxConfigDirs {
		path := filepath.Join(homeDir, dir)
		if _, err := os.Stat(path); err == nil {
			return path, true
		}
	}
	return "", false
}

// platformBrowserMatches reports whether identifier names info by one of its
// Linux desktop file IDs, or names an installed desktop entry with one of
// those IDs (e.g. "Brave Web Browser" for brave-browser.desktop).
func platformBrowserMatches(info browserInfo, identifier string) bool {
	if len(info.LinuxDesktopIDs) == 0 {
		return false
	}
	if slices.Contains(info.LinuxDesktopIDs, identifier) || slices.Contains(info.LinuxDesktopIDs, identifier+".desktop") {
		return true
	}
	entry, ok := findDesktopEntry(loadDesktopEntries(applicationDirs()), identifier)
	return ok && slices.Contains(info.LinuxDesktopIDs, entry.ID)
}
