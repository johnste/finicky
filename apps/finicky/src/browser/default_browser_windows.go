//go:build windows

package browser

// DefaultBrowserName / DefaultBrowserAppType is the out-of-the-box fallback
// browser used when the user hasn't configured one. Microsoft Edge ships on
// every Windows 10/11 machine, so it's the safe default (not Safari, which
// doesn't exist on Windows).
const (
	DefaultBrowserName    = "Microsoft Edge"
	DefaultBrowserAppType = "appName"
)
