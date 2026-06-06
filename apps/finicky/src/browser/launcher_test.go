package browser

import "testing"

func TestFindBrowserInfo(t *testing.T) {
	browsers := []browserInfo{
		{ConfigDirRelative: "Firefox", ID: "org.mozilla.firefox", AppName: "Firefox", Type: "Firefox"},
		{ConfigDirRelative: "Firefox", ID: "org.mozilla.firefoxdeveloperedition", AppName: "Firefox Developer Edition", Type: "Firefox"},
		{ConfigDirRelative: "zen", ID: "app.zen-browser.zen", AppName: "Zen", Type: "Firefox"},
		{ConfigDirRelative: "Google/Chrome", ID: "com.google.Chrome", AppName: "Google Chrome", Type: "Chromium"},
	}

	cases := []struct {
		name       string
		identifier string
		wantID     string // "" means expect no match
	}{
		{"by app name", "Firefox", "org.mozilla.firefox"},
		{"by bundle id", "org.mozilla.firefox", "org.mozilla.firefox"},
		{"by .app path", "/Applications/Firefox.app", "org.mozilla.firefox"},
		{"by .app path with spaces", "/Applications/Firefox Developer Edition.app", "org.mozilla.firefoxdeveloperedition"},
		{"zen by path", "/Applications/Zen.app", "app.zen-browser.zen"},
		{"chromium by path", "/Applications/Google Chrome.app", "com.google.Chrome"},
		{"unknown app name", "Safari", ""},
		{"unknown path", "/Applications/Safari.app", ""},
		{"path basename must not partial-match", "/Applications/Firefoxxx.app", ""},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := findBrowserInfo(browsers, c.identifier)
			if c.wantID == "" {
				if got != nil {
					t.Fatalf("expected no match for %q, got %q", c.identifier, got.ID)
				}
				return
			}
			if got == nil {
				t.Fatalf("expected match %q for %q, got nil", c.wantID, c.identifier)
			}
			if got.ID != c.wantID {
				t.Fatalf("for %q: got %q, want %q", c.identifier, got.ID, c.wantID)
			}
		})
	}
}
