package browser

import (
	"fmt"
	"io"
	"log/slog"
	"os/exec"
	"path/filepath"
	"slices"

	"finicky/util"
)

// browserCommand builds the `open` command that launches the browser.
func browserCommand(config BrowserConfig, openInBackground bool, profileArgs []string, hasProfile bool) (*exec.Cmd, error) {
	var openArgs []string

	if config.AppType == "bundleId" {
		openArgs = []string{"-b", config.Name}
	} else {
		openArgs = []string{"-a", config.Name}
	}

	if openInBackground {
		openArgs = append(openArgs, "-g")
	}

	hasCustomArgs := len(config.Args) > 0

	// Add -n flag if profile is used (required for profile switching)
	if hasProfile {
		openArgs = append(openArgs, "-n")
	}

	// Add --args if we have profile args or custom args
	if hasProfile || hasCustomArgs {
		if !slices.Contains(config.Args, "--args") {
			openArgs = append(openArgs, "--args")
		}
		// Add profile arguments first if present
		if hasProfile {
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

	return exec.Command("open", openArgs...), nil
}

// runBrowserCommand runs `open`, which returns as soon as the browser has
// been asked to open the URL.
func runBrowserCommand(cmd *exec.Cmd) error {
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

// profileRoot returns the directory holding the browser's profiles
// ("Local State" for Chromium browsers, profiles.ini for Firefox).
func profileRoot(info browserInfo) (string, bool) {
	if info.ConfigDirRelative == "" {
		return "", false
	}
	homeDir, err := util.UserHomeDir()
	if err != nil {
		slog.Info("Error getting home directory", "error", err)
		return "", false
	}
	return filepath.Join(homeDir, "Library/Application Support", info.ConfigDirRelative), true
}

// platformBrowserMatches reports whether identifier names info in a
// platform-specific way. macOS browsers are only matched by app name or
// bundle ID.
func platformBrowserMatches(info browserInfo, identifier string) bool {
	return false
}
