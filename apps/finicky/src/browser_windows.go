//go:build windows

package main

import (
	"fmt"
	"syscall"
	"unsafe"
)

// progID is the ProgID the installer registers for Finicky's http/https URL
// associations (see scripts/installer.iss). Windows records the user's default
// browser choice as this ProgID under UserChoice, so isDefaultBrowser must
// compare against it — not the application's display name.
const progID = "FinickyURL"

var (
	advapi32Browser    = syscall.NewLazyDLL("advapi32.dll")
	procRegOpenKeyExB  = advapi32Browser.NewProc("RegOpenKeyExW")
	procRegQueryValueB = advapi32Browser.NewProc("RegQueryValueExW")
	procRegCloseKeyB   = advapi32Browser.NewProc("RegCloseKey")
)

const (
	hkeyCurrentUser = 0x80000001
	keyReadB        = 0x20019
)

// isDefaultBrowser reports whether Finicky is the user's default browser by
// reading the http/https UserChoice ProgIDs. Windows 10+ does not allow apps to
// set this programmatically; the user selects Finicky in Settings (the
// installer can open that page), so there is no setDefaultBrowser counterpart.
func isDefaultBrowser() (bool, error) {
	httpHandler, _ := getDefaultHandlerForURLScheme("http")
	httpsHandler, _ := getDefaultHandlerForURLScheme("https")
	return httpHandler == progID && httpsHandler == progID, nil
}

func getDefaultHandlerForURLScheme(scheme string) (string, error) {
	path := fmt.Sprintf(`SOFTWARE\Microsoft\Windows\Shell\Associations\UrlAssociations\%s\UserChoice`, scheme)
	pathPtr, _ := syscall.UTF16PtrFromString(path)

	var key syscall.Handle
	ret, _, _ := procRegOpenKeyExB.Call(
		hkeyCurrentUser,
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		keyReadB,
		uintptr(unsafe.Pointer(&key)),
	)
	if ret != 0 {
		return "", fmt.Errorf("no default handler found for '%s'", scheme)
	}
	defer procRegCloseKeyB.Call(uintptr(key))

	valueName, _ := syscall.UTF16PtrFromString("ProgId")
	buf := make([]uint16, 256)
	size := uint32(len(buf) * 2)
	ret, _, _ = procRegQueryValueB.Call(
		uintptr(key),
		uintptr(unsafe.Pointer(valueName)),
		0, 0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret != 0 {
		return "", fmt.Errorf("no ProgId found for '%s'", scheme)
	}

	return syscall.UTF16ToString(buf), nil
}
