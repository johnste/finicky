package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreServices
#include <stdlib.h>
#include "main.h"
*/
import "C"

import "finicky/resolver"

// runApp runs the Cocoa app on the main thread until it exits.
func runApp(forceOpenWindow bool, showStatusItem bool, keepRunning bool) {
	C.RunApp(C.bool(forceOpenWindow), C.bool(showStatusItem), C.bool(keepRunning))
}

func setStatusItemError(hasError bool) {
	C.SetStatusItemError(C.bool(hasError))
}

// platformPreflight runs before startup. macOS routes URLs to an already
// running instance itself (see terminateOtherInstances in main_darwin.m), so
// there's nothing to do.
func platformPreflight(devMode bool) (handled bool) {
	return false
}

func platformCleanup() {}

//export HandleURL
func HandleURL(url *C.char, name *C.char, bundleId *C.char, path *C.char, windowTitle *C.char, openInBackground C.bool) {
	var opener resolver.OpenerInfo

	if name != nil && bundleId != nil && path != nil {
		opener = resolver.OpenerInfo{
			Name:     C.GoString(name),
			BundleID: C.GoString(bundleId),
			Path:     C.GoString(path),
		}
		if windowTitle != nil {
			opener.WindowTitle = C.GoString(windowTitle)
		}
	}

	handleURL(C.GoString(url), &opener, bool(openInBackground))
}

//export QueueWindowDisplay
func QueueWindowDisplay(openWindow int32) {
	queueWindowDisplay(openWindow != 0)
}

//export ShowConfigWindow
func ShowConfigWindow() {
	showConfigWindow()
}

//export WindowDidClose
func WindowDidClose() {
	windowDidClose()
}

//export GetCurrentConfigPath
func GetCurrentConfigPath() *C.char {
	if path := currentConfigPath(); path != "" {
		return C.CString(path)
	}
	return nil
}
