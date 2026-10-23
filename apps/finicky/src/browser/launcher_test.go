package browser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLaunchBrowserWindowFocus(t *testing.T) {
	binDir := t.TempDir()
	logPath := filepath.Join(binDir, "command")
	for _, executable := range []string{"open", "osascript"} {
		script := "#!/bin/sh\nprintf '%s\\n' '" + executable + "' > \"$FINICKY_TEST_COMMAND_LOG\"\n" +
			"if [ \"$1\" = '-e' ]; then shift 3; fi\nprintf '%s\\n' \"$@\" >> \"$FINICKY_TEST_COMMAND_LOG\"\n"
		if err := os.WriteFile(filepath.Join(binDir, executable), []byte(script), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("FINICKY_TEST_COMMAND_LOG", logPath)
	for _, test := range []struct {
		name, browser                string
		focus, background, wantFocus bool
		explicitBackground           *bool
	}{
		{"opt in", "Google Chrome", true, false, true, nil},
		{"bundle ID", "com.google.Chrome", true, false, true, nil},
		{"default", "Google Chrome", false, false, false, nil},
		{"background", "Google Chrome", true, true, false, nil},
		{"explicit background", "Google Chrome", true, false, false, boolPointer(true)},
		{"explicit foreground", "Google Chrome", true, true, true, boolPointer(false)},
		{"other browser", "Safari", true, false, false, nil},
	} {
		t.Run(test.name, func(t *testing.T) {
			config := BrowserConfig{Name: test.browser, FocusWindow: test.focus, OpenInBackground: test.explicitBackground,
				Args: []string{"--profile-directory=Profile 2", "https://example.com/?q=a'b"}}
			if test.browser == "com.google.Chrome" {
				config.AppType = "bundleId"
			}
			if err := LaunchBrowser(config, false, test.background); err != nil {
				t.Fatal(err)
			}
			output, err := os.ReadFile(logPath)
			if err != nil {
				t.Fatal(err)
			}
			command := "open"
			if test.wantFocus {
				command = "osascript"
			}
			wantArgs := []string{command, "-a", test.browser}
			if config.AppType == "bundleId" {
				wantArgs[1] = "-b"
			}
			background := test.background
			if test.explicitBackground != nil {
				background = *test.explicitBackground
			}
			if background {
				wantArgs = append(wantArgs, "-g")
			}
			wantArgs = append(append(wantArgs, "--args"), config.Args...)
			if want := strings.Join(wantArgs, "\n") + "\n"; string(output) != want {
				t.Fatalf("want %q, got %q", want, output)
			}
		})
	}
}

func boolPointer(value bool) *bool { return &value }
