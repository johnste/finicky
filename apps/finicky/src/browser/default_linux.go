package browser

// preferredDefaults are tried in order for the fallback browser on Linux,
// which has nothing like Safari that is always installed.
var preferredDefaults = []string{"firefox.desktop", "org.mozilla.firefox.desktop", "chromium.desktop", "google-chrome.desktop"}

// SystemDefault returns the browser used when no configuration says
// otherwise, and its app type for BrowserConfig.AppType: the first installed
// of preferredDefaults, else the first installed browser.
func SystemDefault() (name string, appType string) {
	return systemDefaultFrom(loadDesktopEntries(applicationDirs())), "appName"
}

// SystemDefaultDisplayName returns the fallback browser's name as shown in
// the UI.
func SystemDefaultDisplayName() string {
	name, _ := SystemDefault()
	return name
}

func systemDefaultFrom(entries []desktopEntry) string {
	for _, id := range preferredDefaults {
		for _, e := range entries {
			if e.ID == id && e.Name != "" {
				return e.Name
			}
		}
	}
	if installed := installedBrowsers(entries); len(installed) > 0 {
		return installed[0]
	}
	return "firefox"
}
