package util

import (
	"os"
	"path/filepath"
)

// UserHomeDir returns the user's home directory ($HOME).
func UserHomeDir() (string, error) {
	return os.UserHomeDir()
}

// UserCacheDir returns the user's cache directory ($XDG_CACHE_HOME, or ~/.cache).
func UserCacheDir() (string, error) {
	return os.UserCacheDir()
}

// AppDirName is the name of Finicky's directories inside the user's config
// and cache directories.
const AppDirName = "finicky"

// LogDir returns the directory for Finicky's log files:
// $XDG_STATE_HOME/finicky, or ~/.local/state/finicky.
func LogDir() (string, error) {
	if state := os.Getenv("XDG_STATE_HOME"); state != "" {
		return filepath.Join(state, AppDirName), nil
	}
	home, err := UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "state", AppDirName), nil
}
