package util

/*
#include <stdlib.h>
#include "info.h"
*/
import "C"
import (
	"fmt"
	"path/filepath"
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

// AppDirName is the name of Finicky's directories inside the user's config
// and cache directories.
const AppDirName = "Finicky"

// LogDir returns the directory for Finicky's log files.
func LogDir() (string, error) {
	home, err := UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "Logs", "Finicky"), nil
}
