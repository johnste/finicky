package browser

import (
	"os/exec"
	"strings"
	"testing"
	"time"
)

func TestFindBrowserInfoIdentifiesSafari(t *testing.T) {
	for _, identifier := range []string{"Safari", "com.apple.Safari"} {
		info := findBrowserInfo(identifier)
		if info == nil {
			t.Fatalf("findBrowserInfo(%q) = nil, want the Safari entry", identifier)
		}
		if info.Type != "Safari" {
			t.Errorf("findBrowserInfo(%q).Type = %q, want Safari", identifier, info.Type)
		}
	}

	if info := findBrowserInfo("Not A Browser"); info != nil {
		t.Errorf("findBrowserInfo of an unknown app = %+v, want nil", info)
	}
}

func TestSafariMenuItemTitle(t *testing.T) {
	if got := safariMenuItemTitle("Work"); got != "New Work Window" {
		t.Errorf("safariMenuItemTitle(Work) = %q, want New Work Window", got)
	}
}

func TestSafariOpenURLCommandPassesTheURLAsItsOwnArgument(t *testing.T) {
	url := `https://example.com/?q="quoted"&x=$(echo hi)`
	cmd := safariOpenURLCommand(url)

	want := []string{"open", "-a", "Safari", url}
	if len(cmd.Args) != len(want) {
		t.Fatalf("args = %q, want %q", cmd.Args, want)
	}
	for i := range want {
		if cmd.Args[i] != want[i] {
			t.Errorf("arg %d = %q, want %q", i, cmd.Args[i], want[i])
		}
	}
}

func TestSafariMenuErrorExplainsEachFailure(t *testing.T) {
	if err := safariMenuError(safariMenuOK, "Work"); err != nil {
		t.Errorf("OK status gave error %v", err)
	}

	tests := []struct {
		status int
		want   string
	}{
		{safariMenuNotTrusted, "Accessibility"},
		{safariMenuNotRunning, "not running"},
		{safariMenuNoItem, `no profile named "Work"`},
		{safariMenuTimeout, "in time"},
		{safariMenuNoWindows, "windows"},
		{safariMenuPressFailed, "refused"},
		{99, "unexpected result 99"},
	}
	for _, test := range tests {
		err := safariMenuError(test.status, "Work")
		if err == nil {
			t.Errorf("status %d gave no error", test.status)
			continue
		}
		if !strings.Contains(err.Error(), test.want) {
			t.Errorf("status %d error = %q, want it to mention %q", test.status, err, test.want)
		}
	}
}

func TestSafariPressResultWaitsForAWindowUnlessThePressSurelyFailed(t *testing.T) {
	tests := []struct {
		name    string
		axError int
		want    int
	}{
		{"accepted", axErrorSuccess, safariMenuTimeout},
		// Safari was too busy to answer in time, so the press may still run.
		{"unanswered", axErrorCannotComplete, safariMenuTimeout},
		{"unsupported", axErrorActionUnsupported, safariMenuPressFailed},
	}
	for _, test := range tests {
		if got := safariPressResult(test.axError); got != test.want {
			t.Errorf("%s press: result = %d, want %d", test.name, got, test.want)
		}
	}
}

func TestLaunchSafariProfileDryRunRunsNothing(t *testing.T) {
	if err := launchSafariProfile("https://example.com", "Work", true); err != nil {
		t.Errorf("dry run returned %v, want nil", err)
	}
}

func TestLaunchSafariProfileDoesNotOpenTheURLWhenThePressIsRefused(t *testing.T) {
	defer func(press func(string) int, run func(*exec.Cmd) ([]byte, error)) {
		pressMenuItem, runURLCommand = press, run
	}(pressMenuItem, runURLCommand)

	pressMenuItem = func(string) int { return safariMenuPressFailed }
	ran := false
	runURLCommand = func(*exec.Cmd) ([]byte, error) {
		ran = true
		return nil, nil
	}

	err := launchSafariProfile("https://example.com", "Work", false)
	if err == nil || !strings.Contains(err.Error(), "refused") {
		t.Errorf("error = %v, want the refused press reported", err)
	}
	if ran {
		t.Error("the url was opened although no profile window was")
	}
}

func TestLaunchSafariProfileRetriesARefusedPressWhileSafariStarts(t *testing.T) {
	defer func(press func(string) int, run func(*exec.Cmd) ([]byte, error), launch func() error, delay time.Duration) {
		pressMenuItem, runURLCommand, launchSafari, safariStartupDelay = press, run, launch, delay
	}(pressMenuItem, runURLCommand, launchSafari, safariStartupDelay)

	statuses := []int{safariMenuNotRunning, safariMenuPressFailed, safariMenuOK}
	pressMenuItem = func(string) int {
		status := statuses[0]
		statuses = statuses[1:]
		return status
	}
	launches, runs := 0, 0
	launchSafari = func() error { launches++; return nil }
	runURLCommand = func(*exec.Cmd) ([]byte, error) { runs++; return nil, nil }
	safariStartupDelay = 0

	if err := launchSafariProfile("https://example.com", "Work", false); err != nil {
		t.Fatalf("launch returned %v, want success after the retry", err)
	}
	if launches != 1 || runs != 1 {
		t.Errorf("Safari launched %d times and url opened %d times, want 1 and 1", launches, runs)
	}
	if len(statuses) != 0 {
		t.Errorf("%d presses left unused, want all 3 used", len(statuses))
	}
}

func TestLaunchSafariProfileDoesNotRetryAPressThatMayStillOpenAWindow(t *testing.T) {
	defer func(press func(string) int, run func(*exec.Cmd) ([]byte, error), launch func() error, delay time.Duration) {
		pressMenuItem, runURLCommand, launchSafari, safariStartupDelay = press, run, launch, delay
	}(pressMenuItem, runURLCommand, launchSafari, safariStartupDelay)

	statuses := []int{safariMenuNotRunning, safariMenuTimeout, safariMenuOK}
	pressMenuItem = func(string) int {
		status := statuses[0]
		statuses = statuses[1:]
		return status
	}
	runs := 0
	launchSafari = func() error { return nil }
	runURLCommand = func(*exec.Cmd) ([]byte, error) { runs++; return nil, nil }
	safariStartupDelay = 0

	err := launchSafariProfile("https://example.com", "Work", false)
	if err == nil || !strings.Contains(err.Error(), "in time") {
		t.Errorf("error = %v, want the timeout reported", err)
	}
	if len(statuses) != 1 {
		t.Errorf("menu pressed %d times, want 2: a timed-out press must not be repeated", 3-len(statuses))
	}
	if runs != 0 {
		t.Error("the url was opened although no profile window was")
	}
}
