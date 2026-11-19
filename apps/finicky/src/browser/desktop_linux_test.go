package browser

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseExec(t *testing.T) {
	tests := []struct {
		exec, url string
		want      []string
	}{
		{"/usr/lib/firefox/firefox %u", "https://a.example", []string{"/usr/lib/firefox/firefox", "https://a.example"}},
		{"/usr/lib/firefox/firefox %u", "", []string{"/usr/lib/firefox/firefox"}},
		{"chromium %U --flag", "https://a.example", []string{"chromium", "https://a.example", "--flag"}},
		{`"/opt/My Browser/browser" --name "a \"b\" \\c" %F`, "", []string{"/opt/My Browser/browser", "--name", `a "b" \c`}},
		{"browser --class=%c --icon %i 100%%", "", []string{"browser", "--class=", "--icon", "100%"}},
		{"flatpak run org.mozilla.firefox @@u %u @@", "", []string{"flatpak", "run", "org.mozilla.firefox", "@@u", "@@"}},
	}
	for _, tt := range tests {
		got, err := parseExec(tt.exec, tt.url)
		if err != nil {
			t.Errorf("parseExec(%q): %v", tt.exec, err)
			continue
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Errorf("parseExec(%q, %q) = %q, want %q", tt.exec, tt.url, got, tt.want)
		}
	}

	for _, bad := range []string{`browser "unterminated`, "", "   "} {
		if _, err := parseExec(bad, ""); err == nil {
			t.Errorf("parseExec(%q): expected error", bad)
		}
	}
}

func writeDesktopFile(t *testing.T, dir, name, content string) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}

// fakeApplications sets up XDG data dirs holding a user and a system
// applications directory, and returns them in that order.
func fakeApplications(t *testing.T) (user, system string) {
	t.Helper()
	root := t.TempDir()
	t.Setenv("XDG_DATA_HOME", filepath.Join(root, "home"))
	t.Setenv("XDG_DATA_DIRS", filepath.Join(root, "sys"))
	user = filepath.Join(root, "home", "applications")
	system = filepath.Join(root, "sys", "applications")

	writeDesktopFile(t, system, "firefox.desktop", `[Desktop Entry]
Type=Application
Name=Firefox
Name[sv]=Firefox på svenska
Exec=/usr/lib/firefox/firefox %u
MimeType=text/html;x-scheme-handler/http;x-scheme-handler/https;

[Desktop Action new-private-window]
Name=New Private Window
Exec=/usr/lib/firefox/firefox --private-window %u
`)
	writeDesktopFile(t, system, "brave-browser.desktop", `[Desktop Entry]
Type=Application
Name=Brave Web Browser
Exec=/usr/bin/brave-browser-stable %U
MimeType=x-scheme-handler/https;
`)
	// Overridden by the user's copy below.
	writeDesktopFile(t, system, "chromium.desktop", `[Desktop Entry]
Type=Application
Name=Chromium (system)
Exec=/usr/bin/chromium %U
MimeType=x-scheme-handler/https;
`)
	writeDesktopFile(t, user, "chromium.desktop", `[Desktop Entry]
Type=Application
Name=Chromium
Exec=/usr/bin/chromium --user-flag %U
MimeType=x-scheme-handler/https;
`)
	writeDesktopFile(t, user, "finicky.desktop", `[Desktop Entry]
Type=Application
Name=Finicky
Exec="/home/me/.local/bin/finicky" %u
MimeType=x-scheme-handler/http;x-scheme-handler/https;
`)
	writeDesktopFile(t, system, "hidden-browser.desktop", `[Desktop Entry]
Type=Application
Name=Hidden Browser
Exec=hidden %u
Hidden=true
MimeType=x-scheme-handler/https;
`)
	writeDesktopFile(t, system, "nodisplay-browser.desktop", `[Desktop Entry]
Type=Application
Name=Helper
Exec=helper %u
NoDisplay=true
MimeType=x-scheme-handler/https;
`)
	writeDesktopFile(t, system, "editor.desktop", `[Desktop Entry]
Type=Application
Name=Editor
Exec=editor %F
MimeType=text/plain;
`)
	writeDesktopFile(t, system, "vendor/sub.desktop", `[Desktop Entry]
Type=Application
Name=Nested Browser
Exec=nested %u
MimeType=x-scheme-handler/https;
`)
	return user, system
}

func TestInstalledBrowsers(t *testing.T) {
	fakeApplications(t)
	got := GetInstalledBrowsers()
	want := []string{"Brave Web Browser", "Chromium", "Firefox", "Nested Browser"}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFindDesktopEntry(t *testing.T) {
	fakeApplications(t)
	entries := loadDesktopEntries(applicationDirs())

	tests := []struct{ identifier, wantID string }{
		{"Firefox", "firefox.desktop"},
		{"firefox", "firefox.desktop"},
		{"firefox.desktop", "firefox.desktop"},
		{"Brave Web Browser", "brave-browser.desktop"},
		{"brave-browser-stable", "brave-browser.desktop"}, // executable name
		{"Chromium", "chromium.desktop"},
		{"vendor-sub", "vendor-sub.desktop"},
	}
	for _, tt := range tests {
		entry, ok := findDesktopEntry(entries, tt.identifier)
		if !ok || entry.ID != tt.wantID {
			t.Errorf("findDesktopEntry(%q) = %q, %v; want %q", tt.identifier, entry.ID, ok, tt.wantID)
		}
	}

	// The user's override of chromium.desktop wins over the system one.
	if entry, _ := findDesktopEntry(entries, "Chromium"); entry.Exec != "/usr/bin/chromium --user-flag %U" {
		t.Errorf("got Exec %q, want the user's override", entry.Exec)
	}

	for _, missing := range []string{"Finicky", "finicky", "Hidden Browser", "Safari"} {
		if entry, ok := findDesktopEntry(entries, missing); ok {
			t.Errorf("findDesktopEntry(%q) found %q, want none", missing, entry.ID)
		}
	}
}

func TestBrowserCommand(t *testing.T) {
	fakeApplications(t)

	tests := []struct {
		name        string
		config      BrowserConfig
		profileArgs []string
		want        []string
	}{
		{"by name", BrowserConfig{Name: "Firefox", URL: "https://a.example"}, nil,
			[]string{"/usr/lib/firefox/firefox", "https://a.example"}},
		{"by macOS app name", BrowserConfig{Name: "Brave Browser", URL: "https://a.example"}, nil,
			[]string{"/usr/bin/brave-browser-stable", "https://a.example"}},
		{"by macOS bundle ID", BrowserConfig{Name: "org.mozilla.firefox", AppType: "bundleId", URL: "https://a.example"}, nil,
			[]string{"/usr/lib/firefox/firefox", "https://a.example"}},
		{"with profile", BrowserConfig{Name: "Firefox", URL: "https://a.example"}, []string{"-P", "work"},
			[]string{"/usr/lib/firefox/firefox", "-P", "work", "https://a.example"}},
		{"custom args replace the URL", BrowserConfig{Name: "Chromium", URL: "https://a.example", Args: []string{"--args", "--incognito", "https://b.example"}}, nil,
			[]string{"/usr/bin/chromium", "--user-flag", "--incognito", "https://b.example"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := browserCommand(tt.config, false, tt.profileArgs, tt.profileArgs != nil)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(cmd.Args, tt.want) {
				t.Errorf("got %q, want %q", cmd.Args, tt.want)
			}
		})
	}

	if _, err := browserCommand(BrowserConfig{Name: "Safari", URL: "https://a.example"}, false, nil, false); err == nil {
		t.Error("expected an error for a browser that isn't installed")
	}
}

func TestSystemDefault(t *testing.T) {
	_, system := fakeApplications(t)
	if name, appType := SystemDefault(); name != "Firefox" || appType != "appName" {
		t.Errorf("got %q %q, want Firefox appName", name, appType)
	}

	// Without a preferred browser, the first installed one is used.
	os.Remove(filepath.Join(system, "firefox.desktop"))
	if name := SystemDefaultDisplayName(); name != "Chromium" {
		t.Errorf("got %q, want Chromium", name)
	}
}

func TestProfileRoot(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	info := browserInfo{LinuxConfigDirs: []string{".mozilla/firefox", ".config/mozilla/firefox"}}

	if _, ok := profileRoot(info); ok {
		t.Error("expected no profile root when none of the directories exist")
	}
	want := filepath.Join(home, ".config/mozilla/firefox")
	os.MkdirAll(want, 0755)
	if got, ok := profileRoot(info); !ok || got != want {
		t.Errorf("got %q, %v; want %q", got, ok, want)
	}
}

func TestGetProfilesForBrowser_Firefox(t *testing.T) {
	fakeApplications(t)
	home := t.TempDir()
	t.Setenv("HOME", home)
	dir := filepath.Join(home, ".mozilla/firefox")
	os.MkdirAll(dir, 0755)
	os.WriteFile(filepath.Join(dir, "profiles.ini"), []byte("[Profile0]\nName=default\n\n[Profile1]\nName=work\n"), 0644)

	if got := GetProfilesForBrowser("Firefox"); !reflect.DeepEqual(got, []string{"default", "work"}) {
		t.Errorf("got %q", got)
	}
	args, ok := resolveBrowserProfileArgs("Firefox", "work")
	if !ok || !reflect.DeepEqual(args, []string{"-P", "work"}) {
		t.Errorf("got %q, %v", args, ok)
	}
}
