//go:build darwin

package browser

// DefaultBrowserName / DefaultBrowserAppType is the out-of-the-box fallback
// browser used when the user hasn't configured one. Safari is always present
// on macOS.
const (
	DefaultBrowserName    = "com.apple.Safari"
	DefaultBrowserAppType = "bundleId"
)
