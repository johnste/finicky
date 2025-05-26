package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Foundation
#include "browser.h"
*/
import "C"

import (
	"fmt"
	"unsafe"
)

var bundleId = "se.johnste.finicky"

func isDefaultBrowser() (bool, error) {
	bundleIdHTTP, _ := getDefaultHandlerForURLScheme("http")
	bundleIdHTTPS, _ := getDefaultHandlerForURLScheme("https")
	bundleIdFinicky, _ := getDefaultHandlerForURLScheme("finicky")

	if bundleIdHTTP == bundleIdHTTPS &&
		bundleIdHTTP == bundleId &&
		bundleIdHTTP == bundleIdFinicky {
		return true, nil
	}
	return false, nil
}

func setDefaultBrowser() (bool, error) {
	isDefault, err := isDefaultBrowser()
	if err != nil {
		return false, err
	}
	if isDefault {
		return true, nil
	}

	setDefaultHandlerForURLScheme(bundleId, "http")
	setDefaultHandlerForURLScheme(bundleId, "https")
	setDefaultHandlerForURLScheme(bundleId, "finicky")
	return true, nil
}

func getDefaultHandlerForURLScheme(scheme string) (string, error) {
	// Convert Go string to C string
	cScheme := C.CString(scheme)
	defer C.free(unsafe.Pointer(cScheme))

	// Call the Objective-C function from browse.m
	result := C.getDefaultHandlerForURLScheme(cScheme)
	if result != nil {
		defer C.free(unsafe.Pointer(result))
		return C.GoString(result), nil
	} else {
		return "", fmt.Errorf("no default handler found for '%s'", scheme)
	}
}

func setDefaultHandlerForURLScheme(bundleId string, scheme string) (bool, error) {
	// Convert Go string to C string
	cScheme := C.CString(scheme)
	defer C.free(unsafe.Pointer(cScheme))
	cBundleId := C.CString(bundleId)
	defer C.free(unsafe.Pointer(cBundleId))

	// Call the Objective-C function from browse.m
	result := bool(C.setDefaultHandlerForURLScheme(cBundleId, cScheme))

	return result, nil
}
