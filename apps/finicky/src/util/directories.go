package util

/*
#include <stdlib.h>
#include "info.h"
*/
import "C"
import (
	"fmt"
	"strings"
	"unsafe"
)

// UserHomeDir returns the user's home directory using NSHomeDirectory
func UserHomeDir() (string, error) {
	cDir := C.getNSHomeDirectory()
	defer C.free(unsafe.Pointer(cDir))
	dir := C.GoString(cDir)
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
	if path == home || strings.HasPrefix(path, home+"/") {
		return "~" + path[len(home):]
	}
	return path
}

// UserCacheDir returns the user's cache directory using NSSearchPathForDirectoriesInDomains
func UserCacheDir() (string, error) {
	cDir := C.getNSCacheDirectory()
	if cDir == nil {
		return "", fmt.Errorf("failed to get user cache directory")
	}
	defer C.free(unsafe.Pointer(cDir))
	dir := C.GoString(cDir)
	if dir == "" {
		return "", fmt.Errorf("failed to get user cache directory")
	}
	return dir, nil
}
