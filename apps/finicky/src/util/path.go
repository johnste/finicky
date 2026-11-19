package util

import "strings"

// ShortenPath replaces the user's home directory prefix with ~.
func ShortenPath(path string) string {
	home, err := UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if path == home || strings.HasPrefix(path, home+"/") {
		return "~" + path[len(home):]
	}
	return path
}
