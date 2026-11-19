package util

import (
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GetModifierKeys returns the current state of modifier keys. Reading global
// keyboard state isn't possible for a regular app under Wayland (and needs a
// display connection under X11), so on Linux all keys report as released.
func GetModifierKeys() map[string]bool {
	slog.Debug("Modifier keys are not available on Linux")
	return map[string]bool{
		"shift":    false,
		"option":   false,
		"command":  false,
		"control":  false,
		"capsLock": false,
		"fn":       false,
		"function": false,
	}
}

// GetSystemInfo returns system information. Linux has no separate localized
// computer name, so both fields hold the hostname.
func GetSystemInfo() map[string]string {
	hostname, err := os.Hostname()
	if err != nil {
		slog.Debug("Failed to get hostname", "error", err)
	}
	return map[string]string{
		"localizedName": hostname,
		"name":          hostname,
	}
}

// GetPowerInfo returns power and battery status information from
// /sys/class/power_supply.
func GetPowerInfo() map[string]interface{} {
	return powerInfoFrom("/sys/class/power_supply")
}

func powerInfoFrom(dir string) map[string]interface{} {
	read := func(supply, name string) string {
		b, _ := os.ReadFile(filepath.Join(dir, supply, name))
		return strings.TrimSpace(string(b))
	}

	isConnected, isCharging := false, false
	var percentage interface{}
	hasBattery := false

	supplies, _ := os.ReadDir(dir)
	for _, s := range supplies {
		name := s.Name()
		switch read(name, "type") {
		case "Mains", "USB":
			if read(name, "online") == "1" {
				isConnected = true
			}
		case "Battery":
			if read(name, "scope") == "Device" {
				continue // e.g. a wireless mouse's battery
			}
			if !hasBattery {
				hasBattery = true
				if p, err := strconv.Atoi(read(name, "capacity")); err == nil {
					percentage = p
				}
			}
			if read(name, "status") == "Charging" {
				isCharging = true
			}
		}
	}
	// Without a battery the machine must be on mains power.
	if !hasBattery {
		isConnected = true
	}

	slog.Debug("Power info", "isCharging", isCharging, "isConnected", isConnected, "percentage", percentage)
	return map[string]interface{}{
		"isCharging":  isCharging,
		"isConnected": isConnected,
		"percentage":  percentage,
	}
}

// IsAppRunning checks if a process with the given name is running. The
// identifier is compared, ignoring case, with each process's command name
// and executable file name, so "firefox" or "Firefox" both match Firefox.
func IsAppRunning(identifier string) bool {
	isRunning := isProcessRunning("/proc", identifier)
	slog.Debug("App running info", "identifier", identifier, "isRunning", isRunning)
	return isRunning
}

func isProcessRunning(procDir, identifier string) bool {
	want := strings.ToLower(identifier)
	entries, err := os.ReadDir(procDir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if _, err := strconv.Atoi(e.Name()); err != nil {
			continue
		}
		comm, _ := os.ReadFile(filepath.Join(procDir, e.Name(), "comm"))
		if strings.ToLower(strings.TrimSpace(string(comm))) == want {
			return true
		}
		cmdline, _ := os.ReadFile(filepath.Join(procDir, e.Name(), "cmdline"))
		if exe, _, _ := strings.Cut(string(cmdline), "\x00"); exe != "" && strings.ToLower(filepath.Base(exe)) == want {
			return true
		}
	}
	return false
}
