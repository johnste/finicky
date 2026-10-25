package browser

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework ApplicationServices
#include <stdlib.h>
#include "safari_menu.h"
*/
import "C"

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"
	"unsafe"
)

// Opening a url in a chosen Safari profile.
//
// Safari has no command line or scripting API for profiles: `open -a Safari`
// always lands in whichever profile was last active. The one mechanism it
// does expose is its File menu, which gains a "New <profile> Window" item per
// profile. So the window is opened by pressing that item through the
// Accessibility API, which makes it Safari's active window, and the url is
// then handed over with a plain `open`, which Safari loads into that window.

// safariMenuTimeoutSeconds bounds the wait for the new window to appear.
const safariMenuTimeoutSeconds = 5

// safariStartupDelay lets a freshly launched Safari open its own first window
// before the profile window is requested, so the two are not confused.
const safariStartupDelay = 2 * time.Second

// safariStartupAttempts bounds how often the menu is tried after launching
// Safari, since a slow start can leave its menu bar empty for a while.
const safariStartupAttempts = 3

// Results of pressing a Safari menu item, mirrored from safari_menu.h so that
// Go code, tests included, can name them.
const (
	safariMenuOK         = C.SAFARI_MENU_OK
	safariMenuNotTrusted = C.SAFARI_MENU_NOT_TRUSTED
	safariMenuNotRunning = C.SAFARI_MENU_NOT_RUNNING
	safariMenuNoItem     = C.SAFARI_MENU_NO_ITEM
	safariMenuTimeout    = C.SAFARI_MENU_TIMEOUT
	safariMenuNoWindows  = C.SAFARI_MENU_NO_WINDOWS
)

// safariMenuItemTitle is the File menu item Safari shows for a profile.
func safariMenuItemTitle(profile string) string {
	return "New " + profile + " Window"
}

// safariOpenURLCommand hands url to Safari. The url is its own argument and is
// never interpreted by a shell or a script.
func safariOpenURLCommand(url string) *exec.Cmd {
	return exec.Command("open", "-a", "Safari", url)
}

// safariMenuError explains a failed menu press in terms a user can act on.
func safariMenuError(status int, profile string) error {
	switch status {
	case safariMenuOK:
		return nil
	case safariMenuNotTrusted:
		return fmt.Errorf("Finicky needs Accessibility permission to open Safari profile %q (System Settings > Privacy & Security > Accessibility)", profile)
	case safariMenuNotRunning:
		return fmt.Errorf("Safari is not running, so profile %q could not be opened", profile)
	case safariMenuNoItem:
		return fmt.Errorf("Safari has no %q menu item, so there is no profile named %q", safariMenuItemTitle(profile), profile)
	case safariMenuTimeout:
		return fmt.Errorf("Safari did not open a window for profile %q in time", profile)
	case safariMenuNoWindows:
		return fmt.Errorf("could not read Safari's windows, so profile %q was not opened", profile)
	default:
		return fmt.Errorf("unexpected result %d opening Safari profile %q", status, profile)
	}
}

func pressSafariMenuItem(title string) int {
	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	return int(C.safariOpenWindowFromMenu(cTitle, C.double(safariMenuTimeoutSeconds)))
}

// launchSafariProfile opens url in the named Safari profile. Safari is always
// brought forward: `open` loads the url into Safari's active window, so the new
// profile window has to become it, which rules out opening in the background.
func launchSafariProfile(url string, profile string, dryRun bool) error {
	title := safariMenuItemTitle(profile)
	if dryRun {
		slog.Debug("Would open url in Safari profile (dry run)", "profile", profile, "menuItem", title, "url", url)
		return nil
	}

	slog.Debug("Opening url in Safari profile", "profile", profile, "url", url)
	status := pressSafariMenuItem(title)
	if status == safariMenuNotRunning {
		if err := exec.Command("open", "-a", "Safari").Run(); err != nil {
			return fmt.Errorf("launching Safari failed: %v", err)
		}
		for attempt := 0; attempt < safariStartupAttempts; attempt++ {
			time.Sleep(safariStartupDelay)
			status = pressSafariMenuItem(title)
			if status != safariMenuNotRunning && status != safariMenuNoItem {
				break
			}
		}
	}
	if err := safariMenuError(status, profile); err != nil {
		return err
	}

	output, err := safariOpenURLCommand(url).CombinedOutput()
	if err != nil {
		return fmt.Errorf("opening url in Safari profile %q failed: %v: %s", profile, err, strings.TrimSpace(string(output)))
	}
	return nil
}
