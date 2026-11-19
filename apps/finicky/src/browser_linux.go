package main

import (
	"os/exec"
	"strings"
)

// desktopFileID is the name of the desktop entry that registers Finicky as a
// browser (see scripts/install-linux.sh).
const desktopFileID = "finicky.desktop"

// setDefaultBrowser reports whether Finicky is the default browser. Unlike on
// macOS, where the system asks the user to confirm, it doesn't change the
// default by itself; the install script does that.
func setDefaultBrowser() (bool, error) {
	out, err := exec.Command("xdg-settings", "get", "default-web-browser").Output()
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == desktopFileID, nil
}
