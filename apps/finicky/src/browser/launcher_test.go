package browser

import (
	"bytes"
	"log/slog"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func captureBrowserLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &logs
}

func setBrowserData(t *testing.T, data string) {
	t.Helper()
	previous := browsersJsonData
	browsersJsonData = []byte(data)
	t.Cleanup(func() { browsersJsonData = previous })
}

func fakeOpen(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	capture := filepath.Join(dir, "arguments")
	script := "#!/bin/sh\nprintf '%s\\n' \"$@\" > \"$FINICKY_TEST_OPEN_ARGS\"\nprintf 'fake stdout\\n'\nprintf 'fake stderr\\n' >&2\nexit \"$FINICKY_TEST_OPEN_STATUS\"\n"
	if err := os.WriteFile(filepath.Join(dir, "open"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir)
	t.Setenv("FINICKY_TEST_OPEN_ARGS", capture)
	t.Setenv("FINICKY_TEST_OPEN_STATUS", "0")
	return capture
}

func requireLog(t *testing.T, logs string, marker string, want bool) {
	t.Helper()
	if got := strings.Contains(logs, marker); got != want {
		t.Errorf("log marker %q present = %t, want %t; logs: %s", marker, got, want, logs)
	}
}

func TestLaunchBrowserDryRunAndNoneLogging(t *testing.T) {
	for _, logRequests := range []bool{false, true} {
		t.Run(map[bool]string{false: "quiet", true: "logged"}[logRequests], func(t *testing.T) {
			logs := captureBrowserLogs(t)
			capture := fakeOpen(t)
			setBrowserData(t, `[{"id":"test.browser","app_name":"Test Browser","type":"Unsupported"}]`)
			config := BrowserConfig{Name: "Test Browser", URL: "https://example.test/path"}
			if err := LaunchBrowser(config, true, false, logRequests); err != nil {
				t.Fatal(err)
			}
			requireLog(t, logs.String(), `level=INFO msg="Starting browser" name="Test Browser" url=https://example.test/path`, logRequests)
			requireLog(t, logs.String(), `level=DEBUG msg="Browser found in browsers.json" identifier="Test Browser" type=Unsupported`, logRequests)
			requireLog(t, logs.String(), `level=DEBUG msg="Would run command (dry run)" command=`, logRequests)
			logs.Reset()
			config.AppType = "none"
			if err := LaunchBrowser(config, false, false, logRequests); err != nil {
				t.Fatal(err)
			}
			requireLog(t, logs.String(), `level=INFO msg="AppType is 'none', not launching any browser"`, logRequests)
			requireLog(t, logs.String(), "Starting browser", false)
			if _, err := os.Stat(capture); !os.IsNotExist(err) {
				t.Fatalf("dry run or none executed open: %v", err)
			}
		})
	}
}

func TestLaunchBrowserCommandLoggingAndArguments(t *testing.T) {
	foreground, background := false, true
	for _, test := range []struct {
		name       string
		config     BrowserConfig
		background bool
		args       []string
	}{
		{"name", BrowserConfig{Name: "Test Browser", URL: "https://example.test/a b"}, false, []string{"-a", "Test Browser", "https://example.test/a b"}},
		{"bundle-background", BrowserConfig{Name: "test.browser", AppType: "bundleId", URL: "https://example.test"}, true, []string{"-b", "test.browser", "-g", "https://example.test"}},
		{"foreground-override", BrowserConfig{Name: "Test Browser", OpenInBackground: &foreground, Args: []string{"--flag", "custom value"}, URL: "unused"}, true, []string{"-a", "Test Browser", "--args", "--flag", "custom value"}},
		{"background-override", BrowserConfig{Name: "Test Browser", OpenInBackground: &background, Args: []string{"--args", "--flag", "custom value"}}, false, []string{"-a", "Test Browser", "-g", "--args", "--flag", "custom value"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			logs := captureBrowserLogs(t)
			capture := fakeOpen(t)
			setBrowserData(t, `[]`)
			for _, status := range []string{"0", "7"} {
				t.Setenv("FINICKY_TEST_OPEN_STATUS", status)
				for _, logRequests := range []bool{false, true} {
					logs.Reset()
					err := LaunchBrowser(test.config, false, test.background, logRequests)
					if status == "0" && err != nil {
						t.Fatal(err)
					}
					if status == "7" && (err == nil || err.Error() != "command failed: exit status 7") {
						t.Fatalf("failure = %v, want command failed: exit status 7", err)
					}
					data, err := os.ReadFile(capture)
					if err != nil {
						t.Fatal(err)
					}
					args := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
					if !reflect.DeepEqual(args, test.args) {
						t.Fatalf("arguments = %#v, want %#v", args, test.args)
					}
					requireLog(t, logs.String(), `level=DEBUG msg="Run command" command=`, logRequests)
					requireLog(t, logs.String(), `level=DEBUG msg="Command returned output" output="fake stdout\n"`, logRequests)
					requireLog(t, logs.String(), "fake stdout", logRequests)
					requireLog(t, logs.String(), `level=ERROR msg="Command returned error" error="fake stderr\n"`, true)
				}
			}
		})
	}
}

func TestParseProfilesLogging(t *testing.T) {
	for _, logRequests := range []bool{false, true} {
		t.Run(map[bool]string{false: "quiet", true: "logged"}[logRequests], func(t *testing.T) {
			logs := captureBrowserLogs(t)
			path := filepath.Join(t.TempDir(), "Local State")
			if err := os.WriteFile(path, []byte(`{"profile":{"info_cache":{"Profile 1":{"name":"Work"}}}}`), 0600); err != nil {
				t.Fatal(err)
			}
			for _, profile := range []string{"Work", "Profile 1", "Missing"} {
				logs.Reset()
				got, ok := parseProfiles(path, profile, logRequests)
				want, wantOK := "Profile 1", true
				if profile == "Missing" {
					want, wantOK = "", false
				}
				if got != want || ok != wantOK {
					t.Fatalf("profile %q = %q, %t; want %q, %t", profile, got, ok, want, wantOK)
				}
				requireLog(t, logs.String(), `level=INFO msg="Found profile by name" name=Work path="Profile 1"`, profile == "Work" && logRequests)
				requireLog(t, logs.String(), "trying to find by profile path", profile != "Work" && logRequests)
				requireLog(t, logs.String(), `level=WARN msg="Found profile using profile path" path="Profile 1" name=Work suggestion="Please use the profile name instead"`, profile == "Profile 1")
				requireLog(t, logs.String(), `level=WARN msg="Could not find profile in browser profiles."`, profile == "Missing")
			}
		})
	}
}

func TestProfileDiagnosticsRemain(t *testing.T) {
	for _, logRequests := range []bool{false, true} {
		t.Run(map[bool]string{false: "quiet", true: "logged"}[logRequests], func(t *testing.T) {
			logs := captureBrowserLogs(t)
			for _, test := range []struct{ data, marker string }{
				{"", "Error reading Local State file"},
				{"{", "Error parsing Local State JSON"},
				{`{}`, "Could not find profile section in Local State"},
				{`{"profile":{}}`, "Could not find info_cache in profile section"},
			} {
				path := filepath.Join(t.TempDir(), "Local State")
				if test.data != "" {
					if err := os.WriteFile(path, []byte(test.data), 0600); err != nil {
						t.Fatal(err)
					}
				}
				logs.Reset()
				if got, ok := parseProfiles(path, "Work", logRequests); got != "" || ok {
					t.Fatalf("invalid profile fixture returned %q, %t", got, ok)
				}
				requireLog(t, logs.String(), test.marker, true)
				requireLog(t, logs.String(), "level=INFO", true)
			}
			setBrowserData(t, `{`)
			logs.Reset()
			if args, ok := resolveBrowserProfileArgs("Test Browser", "Work", logRequests); args != nil || ok {
				t.Fatalf("invalid browser data returned %v, %t", args, ok)
			}
			requireLog(t, logs.String(), `level=INFO msg="Error parsing browsers.json"`, true)
			browsersJsonData = []byte(`[{"app_name":"Test Browser","type":"Unsupported"}]`)
			logs.Reset()
			if args, ok := resolveBrowserProfileArgs("Test Browser", "Work", logRequests); args != nil || ok {
				t.Fatalf("unsupported profile returned %v, %t", args, ok)
			}
			requireLog(t, logs.String(), `level=INFO msg="Browser is not a supported browser type, skipping profile detection" identifier="Test Browser"`, true)
			path := filepath.Join(t.TempDir(), "profiles.ini")
			logs.Reset()
			if got, ok := parseFirefoxProfiles(path, "Work"); got != "" || ok {
				t.Fatalf("missing Firefox fixture returned %q, %t", got, ok)
			}
			requireLog(t, logs.String(), "Error reading profiles.ini", true)
			requireLog(t, logs.String(), "Could not find profile in Firefox profiles.", true)
		})
	}
}
