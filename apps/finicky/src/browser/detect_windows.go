//go:build windows

package browser

import (
	"log/slog"
	"sort"
	"strings"
	"syscall"
	"unsafe"
)

var (
	advapi32          = syscall.NewLazyDLL("advapi32.dll")
	procRegOpenKeyEx  = advapi32.NewProc("RegOpenKeyExW")
	procRegEnumKeyEx  = advapi32.NewProc("RegEnumKeyExW")
	procRegQueryValue = advapi32.NewProc("RegQueryValueExW")
	procRegCloseKey   = advapi32.NewProc("RegCloseKey")
)

const (
	hkeyCurrentUser  = 0x80000001
	hkeyLocalMachine = 0x80000002
	keyRead          = 0x20019
)

func GetInstalledBrowsers() []string {
	browsers := getBrowsersFromRegistry()
	sort.Strings(browsers)
	return browsers
}

func getBrowsersFromRegistry() []string {
	const path = `SOFTWARE\Clients\StartMenuInternet`

	// Browsers register their Start Menu entry under HKCU for per-user
	// (non-elevated) installs — which is what the default Chrome/Edge/Firefox
	// installers do — and under HKLM for system-wide installs. Scan both roots
	// and dedupe so per-user installs aren't missed.
	seen := map[string]bool{}
	var names []string
	for _, r := range []struct {
		label string
		root  uintptr
	}{{"HKCU", hkeyCurrentUser}, {"HKLM", hkeyLocalMachine}} {
		found := enumStartMenuBrowsers(r.root, path)
		slog.Debug("Browser registry scan", "root", r.label, "found", len(found))
		for _, name := range found {
			key := strings.ToLower(name)
			if !seen[key] {
				seen[key] = true
				names = append(names, name)
			}
		}
	}
	slog.Debug("Installed browsers detected", "count", len(names), "browsers", names)
	return names
}

func enumStartMenuBrowsers(root uintptr, path string) []string {
	pathPtr, _ := syscall.UTF16PtrFromString(path)

	var key syscall.Handle
	ret, _, _ := procRegOpenKeyEx.Call(
		root,
		uintptr(unsafe.Pointer(pathPtr)),
		0,
		keyRead,
		uintptr(unsafe.Pointer(&key)),
	)
	if ret != 0 {
		// Missing key just means no browsers registered under this root.
		return nil
	}
	defer procRegCloseKey.Call(uintptr(key))

	var names []string
	for i := uint32(0); ; i++ {
		buf := make([]uint16, 256)
		size := uint32(len(buf))
		ret, _, _ := procRegEnumKeyEx.Call(
			uintptr(key),
			uintptr(i),
			uintptr(unsafe.Pointer(&buf[0])),
			uintptr(unsafe.Pointer(&size)),
			0, 0, 0, 0,
		)
		if ret != 0 {
			break
		}
		name := syscall.UTF16ToString(buf[:size])
		displayName := getBrowserDisplayName(key, name)
		if displayName != "" && !strings.EqualFold(displayName, "Finicky") {
			names = append(names, displayName)
		}
	}
	return names
}

func getBrowserDisplayName(parentKey syscall.Handle, subkeyName string) string {
	subPath, _ := syscall.UTF16PtrFromString(subkeyName)

	var subKey syscall.Handle
	ret, _, _ := procRegOpenKeyEx.Call(
		uintptr(parentKey),
		uintptr(unsafe.Pointer(subPath)),
		0,
		keyRead,
		uintptr(unsafe.Pointer(&subKey)),
	)
	if ret != 0 {
		return subkeyName
	}
	defer procRegCloseKey.Call(uintptr(subKey))

	buf := make([]uint16, 256)
	size := uint32(len(buf) * 2)
	emptyStr, _ := syscall.UTF16PtrFromString("")
	ret, _, _ = procRegQueryValue.Call(
		uintptr(subKey),
		uintptr(unsafe.Pointer(emptyStr)),
		0,
		0,
		uintptr(unsafe.Pointer(&buf[0])),
		uintptr(unsafe.Pointer(&size)),
	)
	if ret == 0 && size > 2 {
		return syscall.UTF16ToString(buf)
	}
	return subkeyName
}
