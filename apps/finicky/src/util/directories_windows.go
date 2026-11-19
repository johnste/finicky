//go:build windows

package util

import (
	"fmt"
	"os"
	"strings"
)

func UserHomeDir() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil || dir == "" {
		return "", fmt.Errorf("failed to get user home directory: %w", err)
	}
	return dir, nil
}

func ShortenPath(path string) string {
	home, err := UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if path == home || strings.HasPrefix(path, home+`\`) || strings.HasPrefix(path, home+"/") {
		return "~" + path[len(home):]
	}
	return path
}

func UserCacheDir() (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil || dir == "" {
		return "", fmt.Errorf("failed to get user cache directory: %w", err)
	}
	return dir, nil
}
