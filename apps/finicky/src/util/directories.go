package util

/*
#include "info.h"
*/
import "C"
import (
	"fmt"
	"strings"
)

// UserHomeDir returns the user's home directory using NSHomeDirectory
func UserHomeDir() (string, error) {
	dir := C.GoString(C.getNSHomeDirectory())
	if dir == "" {
		return "", fmt.Errorf("failed to get user home directory")
	}
	return dir, nil
}

// ShortenPath replaces the user's home directory prefix with ~.
func ShortenPath(path string) string {
	home, err := UserHomeDir()
	if err != nil || home == "" {
		return path
	}
	if strings.HasPrefix(path, home) {
		return "~" + path[len(home):]
	}
	return path
}

// UserCacheDir returns the user's cache directory using NSSearchPathForDirectoriesInDomains
func UserCacheDir() (string, error) {
	dir := C.GoString(C.getNSCacheDirectory())
	if dir == "" {
		return "", fmt.Errorf("failed to get user cache directory")
	}
	return dir, nil
}