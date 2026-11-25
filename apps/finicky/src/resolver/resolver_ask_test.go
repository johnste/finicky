package resolver_test

import (
	"testing"

	. "finicky/resolver"
)

// Proves the whole config path for appType "ask": the JS config API must carry
// the picker's options through to the Go BrowserConfig with their profiles and
// labels intact, since the picker has nothing to show otherwise.
func TestResolveURL_AskDefaultBrowser(t *testing.T) {
	vm := jsVM(t, `({
		defaultBrowser: {
			name: "ask",
			appType: "ask",
			title: "Open where?",
			options: [
				{ name: "Google Chrome", profile: "Profile 2", title: "Work" },
				"Google Chrome:Default"
			]
		}
	})`)

	result, err := ResolveURL(vm, "https://example.com", nil, false)
	if err != nil {
		t.Fatal(err)
	}

	if result.AppType != "ask" {
		t.Errorf("appType = %q, want %q", result.AppType, "ask")
	}
	if result.Title != "Open where?" {
		t.Errorf("title = %q, want %q", result.Title, "Open where?")
	}
	if len(result.Options) != 2 {
		t.Fatalf("got %d options, want 2", len(result.Options))
	}

	if got := result.Options[0]; got.Name != "Google Chrome" || got.Profile != "Profile 2" || got.Title != "Work" {
		t.Errorf("option 0 = %+v, want Google Chrome / Profile 2 / Work", got)
	}
	// The "name:profile" shorthand must expand the same way it does elsewhere.
	if got := result.Options[1]; got.Name != "Google Chrome" || got.Profile != "Default" || got.AppType != "appName" {
		t.Errorf("option 1 = %+v, want Google Chrome / Default / appName", got)
	}
}

// A handler resolving to "ask" must behave the same as a default browser does.
func TestResolveURL_AskHandler(t *testing.T) {
	vm := jsVM(t, `({
		defaultBrowser: "Safari",
		handlers: [
			{
				match: "*example.com/*",
				browser: { name: "ask", appType: "ask", options: ["Safari", "Firefox"] }
			}
		]
	})`)

	result, err := ResolveURL(vm, "https://example.com/page", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.AppType != "ask" {
		t.Fatalf("appType = %q, want %q", result.AppType, "ask")
	}
	if len(result.Options) != 2 || result.Options[0].Name != "Safari" || result.Options[1].Name != "Firefox" {
		t.Errorf("options = %+v, want Safari then Firefox", result.Options)
	}

	// A url the handler does not match must be untouched by any of this.
	other, err := ResolveURL(vm, "https://other.test/page", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if other.AppType == "ask" {
		t.Error("unmatched url resolved to the picker")
	}
	if other.Name != "Safari" {
		t.Errorf("unmatched url = %q, want Safari", other.Name)
	}
}
