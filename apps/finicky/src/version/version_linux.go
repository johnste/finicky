package version

// appVersion is set via ldflags during build (see scripts/build-linux.sh),
// since there's no app bundle to read it from.
var appVersion = "dev"

// GetCurrentVersion returns the app version the binary was built with.
func GetCurrentVersion() string {
	return appVersion
}
