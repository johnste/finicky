package util

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreServices -framework IOKit
#include <stdlib.h>
#include "info.h"
// PowerInfo struct and getPowerInfo function for power status
*/
import "C"
import "log/slog"

// GetModifierKeys returns the current state of modifier keys
func GetModifierKeys() map[string]bool {
	keys := C.getModifierKeys()
	result := map[string]bool{
		"shift":    bool(keys.shift),
		"option":   bool(keys.option),
		"command":  bool(keys.command),
		"control":  bool(keys.control),
		"capsLock": bool(keys.capsLock),
		"fn":       bool(keys.fn),
		"function":	bool(keys.fn),
	}
	args := []any{}
	for k, v := range result {
		if k == "function" {
			continue
		}
		args = append(args, k, v)
	}
	slog.Debug("Modifier keys state", args...)
	return result
}

// GetSystemInfo returns system information
func GetSystemInfo() map[string]string {
	info := C.getSystemInfo()
	return map[string]string{
		"localizedName": C.GoString(info.localizedName),
		"name":          C.GoString(info.name),
	}
}

// GetPowerInfo returns power and battery status information
func GetPowerInfo() map[string]interface{} {
	info := C.getPowerInfo()

	percentage := int(info.percentage)

	if percentage == -1 {
		slog.Debug("Power info", "isCharging", info.isCharging, "isConnected", info.isConnected, "percentage", nil)
		return map[string]interface{}{
			"isCharging":   bool(info.isCharging),
			"isConnected": bool(info.isConnected),
			"percentage":  nil,
		}
	}

	slog.Debug("Power info", "isCharging", info.isCharging, "isConnected", info.isConnected, "percentage", info.percentage)
	return map[string]interface{}{
		"isCharging":   bool(info.isCharging),
		"isConnected": bool(info.isConnected),
		"percentage":  percentage,
	}
}
