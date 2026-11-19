package browser

// SystemDefault returns the browser used when no configuration says
// otherwise, and its app type for BrowserConfig.AppType.
func SystemDefault() (name string, appType string) {
	return "com.apple.Safari", "bundleId"
}

// SystemDefaultDisplayName returns the fallback browser's name as shown in
// the UI.
func SystemDefaultDisplayName() string {
	return "Safari"
}
