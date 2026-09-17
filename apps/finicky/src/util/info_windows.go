//go:build windows

package util

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	procGetAsyncKeyState = user32.NewProc("GetAsyncKeyState")
	procGetWindowTextW   = user32.NewProc("GetWindowTextW")

	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW          = kernel32.NewProc("Process32FirstW")
	procProcess32NextW           = kernel32.NewProc("Process32NextW")
	procGetSystemPowerStatus     = kernel32.NewProc("GetSystemPowerStatus")
)

const (
	vkShift   = 0x10
	vkControl = 0x11
	vkMenu    = 0x12 // Alt/Option
	vkLWin    = 0x5B // Left Windows key (maps to Command)
	vkCapital = 0x14
)

func GetModifierKeys() map[string]bool {
	result := map[string]bool{
		"shift":    isKeyDown(vkShift),
		"option":   isKeyDown(vkMenu),
		"command":  isKeyDown(vkLWin),
		"control":  isKeyDown(vkControl),
		"capsLock": isKeyDown(vkCapital),
		"fn":       false, // no direct equivalent on Windows
		"function": false,
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

func isKeyDown(vk int) bool {
	ret, _, _ := procGetAsyncKeyState.Call(uintptr(vk))
	return ret&0x8000 != 0
}

func GetSystemInfo() map[string]string {
	name, err := os.Hostname()
	if err != nil {
		name = ""
	}
	return map[string]string{
		"localizedName": name,
		"name":          name,
	}
}

type systemPowerStatus struct {
	ACLineStatus        byte
	BatteryFlag         byte
	BatteryLifePercent  byte
	SystemStatusFlag    byte
	BatteryLifeTime     uint32
	BatteryFullLifeTime uint32
}

func GetPowerInfo() map[string]interface{} {
	var status systemPowerStatus
	ret, _, _ := procGetSystemPowerStatus.Call(uintptr(unsafe.Pointer(&status)))
	if ret == 0 {
		slog.Debug("Power info", "isCharging", false, "isConnected", false, "percentage", nil)
		return map[string]interface{}{
			"isCharging":  false,
			"isConnected": false,
			"percentage":  nil,
		}
	}

	isConnected := status.ACLineStatus == 1
	isCharging := status.BatteryFlag&0x08 != 0

	var percentage interface{}
	if status.BatteryLifePercent <= 100 {
		percentage = int(status.BatteryLifePercent)
	}

	slog.Debug("Power info", "isCharging", isCharging, "isConnected", isConnected, "percentage", percentage)
	return map[string]interface{}{
		"isCharging":  isCharging,
		"isConnected": isConnected,
		"percentage":  percentage,
	}
}

const thSnapshot = 0x00000002 // TH32CS_SNAPPROCESS

type processEntry32W struct {
	Size            uint32
	CntUsage        uint32
	ProcessID       uint32
	DefaultHeapID   uintptr
	ModuleID        uint32
	CntThreads      uint32
	ParentProcessID uint32
	PriClassBase    int32
	Flags           uint32
	ExeFile         [260]uint16
}

func IsAppRunning(identifier string) bool {
	handle, _, _ := procCreateToolhelp32Snapshot.Call(thSnapshot, 0)
	if handle == ^uintptr(0) {
		return false
	}
	defer syscall.CloseHandle(syscall.Handle(handle))

	var entry processEntry32W
	entry.Size = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32FirstW.Call(handle, uintptr(unsafe.Pointer(&entry)))
	if ret == 0 {
		return false
	}

	target := strings.ToLower(identifier)
	for {
		name := strings.ToLower(syscall.UTF16ToString(entry.ExeFile[:]))
		// Match against exe name with or without .exe suffix
		if name == target || name == target+".exe" ||
			strings.TrimSuffix(name, ".exe") == target {
			slog.Debug("App running info", "identifier", identifier, "isRunning", true)
			return true
		}

		entry.Size = uint32(unsafe.Sizeof(entry))
		ret, _, _ = procProcess32NextW.Call(handle, uintptr(unsafe.Pointer(&entry)))
		if ret == 0 {
			break
		}
	}

	slog.Debug("App running info", "identifier", identifier, "isRunning", false)
	return false
}

// GetForegroundWindowTitle returns the title of the currently focused window.
func GetForegroundWindowTitle() string {
	procGetForegroundWindow := user32.NewProc("GetForegroundWindow")
	hwnd, _, _ := procGetForegroundWindow.Call()
	if hwnd == 0 {
		return ""
	}

	buf := make([]uint16, 256)
	ret, _, _ := procGetWindowTextW.Call(hwnd, uintptr(unsafe.Pointer(&buf[0])), 256)
	if ret == 0 {
		return ""
	}
	return syscall.UTF16ToString(buf)
}

// LogDir returns the platform-appropriate log directory for Finicky.
func LogDir() (string, error) {
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return "", fmt.Errorf("APPDATA environment variable not set")
	}
	return appData + `\Finicky\Logs`, nil
}
