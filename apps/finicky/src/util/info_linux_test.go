package util

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func writeFiles(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, content := range files {
		path := filepath.Join(root, name)
		os.MkdirAll(filepath.Dir(path), 0755)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
	}
}

func TestPowerInfoFrom(t *testing.T) {
	tests := []struct {
		name  string
		files map[string]string
		want  map[string]interface{}
	}{
		{"laptop charging", map[string]string{
			"AC/type": "Mains", "AC/online": "1",
			"BAT0/type": "Battery", "BAT0/scope": "", "BAT0/capacity": "57", "BAT0/status": "Charging",
		}, map[string]interface{}{"isConnected": true, "isCharging": true, "percentage": 57}},
		{"laptop on battery, mouse battery ignored", map[string]string{
			"AC/type": "Mains", "AC/online": "0",
			"hidpp_battery_0/type": "Battery", "hidpp_battery_0/scope": "Device", "hidpp_battery_0/capacity": "5",
			"BAT1/type": "Battery", "BAT1/capacity": "80", "BAT1/status": "Discharging",
		}, map[string]interface{}{"isConnected": false, "isCharging": false, "percentage": 80}},
		{"desktop", map[string]string{}, map[string]interface{}{"isConnected": true, "isCharging": false, "percentage": nil}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			writeFiles(t, dir, tt.files)
			if got := powerInfoFrom(dir); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsProcessRunning(t *testing.T) {
	dir := t.TempDir()
	writeFiles(t, dir, map[string]string{
		"1/comm":       "systemd\n",
		"1/cmdline":    "/sbin/init\x00splash\x00",
		"42/comm":      "GeckoMain\n",
		"42/cmdline":   "/usr/lib/firefox/firefox\x00-P\x00work\x00",
		"self/comm":    "ignored\n",
		"notapid/comm": "ignored\n",
	})

	for _, name := range []string{"firefox", "Firefox", "GeckoMain", "systemd", "init"} {
		if !isProcessRunning(dir, name) {
			t.Errorf("%q: expected running", name)
		}
	}
	for _, name := range []string{"chromium", "ignored", "work", ""} {
		if isProcessRunning(dir, name) {
			t.Errorf("%q: expected not running", name)
		}
	}
}
