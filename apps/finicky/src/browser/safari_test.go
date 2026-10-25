package browser

import (
	"strings"
	"testing"
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

func TestLaunchSafariProfileDryRunRunsNothing(t *testing.T) {
	if err := launchSafariProfile("https://example.com", "Work", true); err != nil {
		t.Errorf("dry run returned %v, want nil", err)
	}
}
