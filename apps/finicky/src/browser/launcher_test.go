package browser

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"finicky/util"
)

func TestLaunchBrowserWindowFocus(t *testing.T) {
	for _, test := range []struct {
		name       string
		browser    string
		background bool
		wantFocus  bool
	}{
		{"opt in", `{"name":"Google Chrome","profile":"Work","focusWindow":true,"url":"https://example.com/"}`, false, true},
		{"bundle ID", `{"name":"com.google.Chrome","appType":"bundleId","profile":"Work","focusWindow":true,"url":"https://example.com/"}`, false, true},
		{"default", `{"name":"Google Chrome","url":"https://example.com/"}`, false, false},
		{"disabled", `{"name":"Google Chrome","focusWindow":false,"url":"https://example.com/"}`, false, false},
		{"background", `{"name":"Google Chrome","profile":"Work","focusWindow":true,"url":"https://example.com/"}`, true, false},
		{"explicit background", `{"name":"Google Chrome","profile":"Work","focusWindow":true,"openInBackground":true,"url":"https://example.com/"}`, false, false},
		{"explicit foreground", `{"name":"Google Chrome","profile":"Work","focusWindow":true,"openInBackground":false,"url":"https://example.com/"}`, true, true},
		{"other browser", `{"name":"Safari","focusWindow":true,"url":"https://example.com/"}`, false, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			profileDir := t.TempDir()
			if err := os.WriteFile(filepath.Join(profileDir, "Local State"), []byte(`{"profile":{"info_cache":{"Profile 2":{"name":"Work"}}}}`), 0600); err != nil {
				t.Fatal(err)
			}
			homeDir, err := util.UserHomeDir()
			if err != nil {
				t.Fatal(err)
			}
			relativeProfileDir, err := filepath.Rel(filepath.Join(homeDir, "Library/Application Support"), profileDir)
			if err != nil {
				t.Fatal(err)
			}
			originalBrowserData := browsersJsonData
			t.Cleanup(func() { browsersJsonData = originalBrowserData })
			browsersJsonData, err = json.Marshal([]browserInfo{{
				AppName: "Google Chrome", ID: "com.google.Chrome", Type: "Chromium", ConfigDirRelative: relativeProfileDir,
			}})
			if err != nil {
				t.Fatal(err)
			}
			binDir := t.TempDir()
			logPath := filepath.Join(binDir, "commands")
			for _, executable := range []string{"open", "osascript"} {
				script := "#!/bin/sh\nprintf '%s\\n' '" + executable + "' \"$@\" >> \"$FINICKY_TEST_COMMAND_LOG\"\n"
				if err := os.WriteFile(filepath.Join(binDir, executable), []byte(script), 0700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
			t.Setenv("FINICKY_TEST_COMMAND_LOG", logPath)
			var config BrowserConfig
			if err := json.Unmarshal([]byte(test.browser), &config); err != nil {
				t.Fatal(err)
			}
			if err := LaunchBrowser(config, false, test.background); err != nil {
				t.Fatal(err)
			}
			output, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			wantCommand := "open\n"
			if test.wantFocus {
				wantCommand = "osascript\n"
			}
			if !strings.HasPrefix(string(output), wantCommand) {
				t.Fatalf("want %q launcher, got %q", strings.TrimSpace(wantCommand), string(output))
			}
			wantBackground := test.background
			if config.OpenInBackground != nil {
				wantBackground = *config.OpenInBackground
			}
			if strings.Contains(string(output), "\n-g\n") != wantBackground {
				t.Fatalf("background flag differs from requested behavior: %s", output)
			}
			if config.Profile != "" && (!strings.Contains(string(output), "\n-n\n") || !strings.Contains(string(output), "\n--args\n")) {
				t.Fatalf("profile launch flags missing from command: %s", output)
			}
			if config.Profile != "" && !strings.Contains(string(output), "--profile-directory=Profile 2\n") {
				t.Fatalf("resolved profile missing from command: %s", output)
			}
			if !strings.Contains(string(output), "https://example.com/\n") {
				t.Fatalf("URL missing from command: %s", output)
			}
		})
	}
}

func TestLaunchBrowserFocusDryRun(t *testing.T) {
	binDir := t.TempDir()
	t.Setenv("PATH", binDir)
	if err := LaunchBrowser(BrowserConfig{Name: "Google Chrome", FocusWindow: true, URL: "https://example.com/"}, true, false); err != nil {
		t.Fatal(err)
	}
}

func TestLaunchBrowserFocusDisabledApp(t *testing.T) {
	if err := LaunchBrowser(BrowserConfig{AppType: "none", FocusWindow: true}, false, false); err != nil {
		t.Fatal(err)
	}
}
