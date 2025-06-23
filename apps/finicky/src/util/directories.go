package util

/*
#include "info.h"
*/
import "C"

// UserHomeDir returns the user's home directory using NSHomeDirectory
func UserHomeDir() string {
	return C.GoString(C.getNSHomeDirectory())
}

// UserCacheDir returns the user's cache directory using NSSearchPathForDirectoriesInDomains
func UserCacheDir() string {
	return C.GoString(C.getNSCacheDirectory())
}