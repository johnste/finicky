//go:build windows

package version

import (
	"fmt"
	"log/slog"
	"os"
	"unsafe"

	"golang.org/x/sys/windows"
)

// getCurrentVersionPlatform reads the product version from the executable's
// embedded PE version resource (see versioninfo.json /
// resource_windows_amd64.syso). Windows has no `defaults`/Info.plist
// equivalent, so we read the resource directly instead of shelling out to a
// macOS-only tool. Fallback only — the ldflags-injected version wins.
func getCurrentVersionPlatform() string {
	exePath, err := os.Executable()
	if err != nil {
		slog.Error("Error getting executable path", "error", err)
		return "dev"
	}

	var zero windows.Handle
	size, err := windows.GetFileVersionInfoSize(exePath, &zero)
	if err != nil || size == 0 {
		slog.Debug("No version resource in executable", "error", err)
		return "dev"
	}

	info := make([]byte, size)
	if err := windows.GetFileVersionInfo(exePath, 0, size, unsafe.Pointer(&info[0])); err != nil {
		slog.Debug("Failed to read version resource", "error", err)
		return "dev"
	}

	var fixed *windows.VS_FIXEDFILEINFO
	fixedLen := uint32(unsafe.Sizeof(*fixed))
	if err := windows.VerQueryValue(unsafe.Pointer(&info[0]), `\`, unsafe.Pointer(&fixed), &fixedLen); err != nil || fixed == nil {
		slog.Debug("Failed to query fixed file info", "error", err)
		return "dev"
	}

	major := (fixed.FileVersionMS >> 16) & 0xffff
	minor := fixed.FileVersionMS & 0xffff
	patch := (fixed.FileVersionLS >> 16) & 0xffff
	return fmt.Sprintf("%d.%d.%d", major, minor, patch)
}
