package util

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreServices
#include <stdlib.h>
#include "info.h"
*/
import "C"

// GetModifierKeys returns the current state of modifier keys
func GetModifierKeys() map[string]bool {
	keys := C.getModifierKeys()
	return map[string]bool{
		"shift":    bool(keys.shift),
		"option":   bool(keys.option),
		"command":  bool(keys.command),
		"control":  bool(keys.control),
		"capsLock": bool(keys.capsLock),
		"fn":       bool(keys.fn),
		"function":	bool(keys.fn),
	}
}

// GetSystemInfo returns system information
func GetSystemInfo() map[string]string {
	info := C.getSystemInfo()
	return map[string]string{
		"localizedName": C.GoString(info.localizedName),
		"name":          C.GoString(info.name),
	}
}