package browser

import "sort"

// GetInstalledBrowsers returns the names of all applications whose desktop
// entries handle https:// URLs, excluding Finicky itself.
func GetInstalledBrowsers() []string {
	return installedBrowsers(loadDesktopEntries(applicationDirs()))
}

func installedBrowsers(entries []desktopEntry) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, e := range entries {
		if e.NoShow || !e.handlesHTTPS() || e.isFinicky() || e.Name == "" || seen[e.Name] {
			continue
		}
		seen[e.Name] = true
		result = append(result, e.Name)
	}
	sort.Strings(result)
	return result
}
