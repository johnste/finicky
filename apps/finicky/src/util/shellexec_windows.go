//go:build windows

package util

import (
	"fmt"
	"syscall"
	"unsafe"
)

var (
	shell32Util       = syscall.NewLazyDLL("shell32.dll")
	procShellExecuteW = shell32Util.NewProc("ShellExecuteW")
)

// OpenURLDefault opens a URL with the system's default handler via ShellExecuteW.
//
// Unlike `cmd /c start <url>`, the URL is handed to the shell as a single
// wide-string argument and is never parsed by cmd.exe, so shell metacharacters
// in an attacker-controlled URL (&, |, ^, >, quotes, ...) cannot break out and
// inject commands. Use this instead of shelling out to cmd for URL opening.
func OpenURLDefault(url string) error {
	verb, err := syscall.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	target, err := syscall.UTF16PtrFromString(url)
	if err != nil {
		return err
	}
	const swShowNormal = 1
	// ShellExecuteW(hwnd, lpVerb, lpFile, lpParameters, lpDirectory, nShowCmd)
	ret, _, callErr := procShellExecuteW.Call(
		0,
		uintptr(unsafe.Pointer(verb)),
		uintptr(unsafe.Pointer(target)),
		0,
		0,
		swShowNormal,
	)
	// ShellExecuteW returns a value greater than 32 on success.
	if ret <= 32 {
		if callErr != nil && callErr != syscall.Errno(0) {
			return fmt.Errorf("ShellExecuteW failed: %w", callErr)
		}
		return fmt.Errorf("ShellExecuteW failed with code %d", ret)
	}
	return nil
}
