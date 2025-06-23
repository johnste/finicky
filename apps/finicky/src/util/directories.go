package util

/*
#include "info.h"
*/
import "C"
import "fmt"

// UserHomeDir returns the user's home directory using NSHomeDirectory
func UserHomeDir() (string, error) {
	dir := C.GoString(C.getNSHomeDirectory())
	if dir == "" {
		return "", fmt.Errorf("failed to get user home directory")
	}
	return dir, nil
}

// UserCacheDir returns the user's cache directory using NSSearchPathForDirectoriesInDomains
func UserCacheDir() (string, error) {
	dir := C.GoString(C.getNSCacheDirectory())
	if dir == "" {
		return "", fmt.Errorf("failed to get user cache directory")
	}
	return dir, nil
}