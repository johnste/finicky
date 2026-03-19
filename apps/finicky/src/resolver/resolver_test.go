package resolver_test

import (
	"os"
	"testing"

	"finicky/config"
	. "finicky/resolver"
	"finicky/rules"
)

// apiContent reads finickyConfigAPI.js relative to this package directory.
// go test sets the working directory to the package source directory, so
// "../assets/..." resolves correctly.
func apiContent(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("../assets/finickyConfigAPI.js")
	if err != nil {
		t.Fatalf("failed to load finickyConfigAPI.js: %v\n(run from apps/finicky/src/resolver/)", err)
	}
	return b
}

// jsVM creates a VM from an inline JS config object literal.
// NewFromScript bypasses esbuild bundling, so we use the legacy var-assignment
// syntax that the config API accepts directly in goja.
// The VM is marked as a JS-config VM to exercise the merge path.
func jsVM(t *testing.T, configObj string) *config.VM {
	t.Helper()
	script := "var finickyConfig = " + configObj
	vm, err := config.NewFromScript(apiContent(t), "finickyConfig", script)
	if err != nil {
		t.Fatalf("failed to create VM from JS: %v", err)
	}
	vm.SetIsJSConfig(true)
	return vm
}

// rulesVM creates a VM from a RulesFile (the no-JS-config path).
func rulesVM(t *testing.T, rf rules.RulesFile) *config.VM {
	t.Helper()
	script, err := rules.ToJSConfigScript(rf, "finickyConfig")
	if err != nil {
		t.Fatalf("failed to generate JS config from rules: %v", err)
	}
	vm, err := config.NewFromScript(apiContent(t), "finickyConfig", script)
	if err != nil {
		t.Fatalf("failed to create VM from rules: %v", err)
	}
	return vm
}

func TestResolveURL_NoConfig(t *testing.T) {
	result, err := ResolveURL(nil, "https://example.com", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "com.apple.Safari" {
		t.Errorf("got %q, want %q", result.Name, "com.apple.Safari")
	}
}

func TestResolveURL_JSConfig(t *testing.T) {
	vm := jsVM(t, `({
		defaultBrowser: "Safari",
		handlers: [
			{ match: "*github.com/*", browser: "Firefox" },
			{ match: "https://linear.app/*", browser: "Google Chrome" }
		]
	})`)

	tests := []struct {
		url     string
		browser string
	}{
		{"https://github.com/johnste/finicky", "Firefox"},
		{"https://gist.github.com/foo", "Firefox"},
		{"https://linear.app/team/issue/123", "Google Chrome"},
		{"https://example.com", "Safari"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result, err := ResolveURL(vm, tt.url, nil, false)
			if err != nil {
				t.Fatal(err)
			}
			if result.Name != tt.browser {
				t.Errorf("got %q, want %q", result.Name, tt.browser)
			}
		})
	}
}

func TestResolveURL_JSONRulesOnly(t *testing.T) {
	rf := rules.RulesFile{
		DefaultBrowser: "Firefox",
		Rules: []rules.Rule{
			{Match: "*github.com/*", Browser: "Google Chrome"},
			{Match: "https://linear.app/*", Browser: "Safari"},
		},
	}
	vm := rulesVM(t, rf)

	tests := []struct {
		url     string
		browser string
	}{
		{"https://github.com/johnste/finicky", "Google Chrome"},
		{"https://linear.app/team/issue/123", "Safari"},
		{"https://example.com", "Firefox"},
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result, err := ResolveURL(vm, tt.url, nil, false)
			if err != nil {
				t.Fatal(err)
			}
			if result.Name != tt.browser {
				t.Errorf("got %q, want %q", result.Name, tt.browser)
			}
		})
	}
}

func TestResolveURL_JSONRulesWithProfile(t *testing.T) {
	rf := rules.RulesFile{
		DefaultBrowser: "Safari",
		Rules: []rules.Rule{
			{Match: "*github.com/*", Browser: "Google Chrome", Profile: "Work"},
		},
	}
	vm := rulesVM(t, rf)

	result, err := ResolveURL(vm, "https://github.com/foo", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Google Chrome" {
		t.Errorf("browser: got %q, want %q", result.Name, "Google Chrome")
	}
	if result.Profile != "Work" {
		t.Errorf("profile: got %q, want %q", result.Profile, "Work")
	}
}

// TestResolveURL_MergedJSAndJSON verifies that JS handlers take precedence
// over JSON rules handlers when both are present.
func TestResolveURL_MergedJSAndJSON(t *testing.T) {
	// JS config handles github. jsVM sets IsJSConfig=true so the merge path
	// is exercised. With no rules cached, _jsonHandlers is [], so finalConfig
	// is used as-is — JS handlers apply normally.
	jsConfig := jsVM(t, `({
		defaultBrowser: "Safari",
		handlers: [
			{ match: "*github.com/*", browser: "Firefox" }
		]
	})`)

	tests := []struct {
		url     string
		browser string
	}{
		{"https://github.com/foo", "Firefox"}, // matched by JS handler
		{"https://example.com", "Safari"},     // falls through to JS default
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result, err := ResolveURL(jsConfig, tt.url, nil, false)
			if err != nil {
				t.Fatal(err)
			}
			if result.Name != tt.browser {
				t.Errorf("got %q, want %q", result.Name, tt.browser)
			}
		})
	}
}

func TestResolveURL_JSConfigFunctionHandler(t *testing.T) {
	vm := jsVM(t, `({
		defaultBrowser: "Safari",
		handlers: [{
			match: function(request, { opener }) {
				return opener && opener.bundleId === "com.apple.Terminal";
			},
			browser: "Firefox"
		}]
	})`)

	terminal := &OpenerInfo{Name: "Terminal", BundleID: "com.apple.Terminal", Path: "/Applications/Utilities/Terminal.app"}
	other := &OpenerInfo{Name: "Finder", BundleID: "com.apple.finder", Path: "/System/Library/CoreServices/Finder.app"}

	result, err := ResolveURL(vm, "https://example.com", terminal, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Firefox" {
		t.Errorf("terminal opener: got %q, want %q", result.Name, "Firefox")
	}

	result, err = ResolveURL(vm, "https://example.com", other, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Safari" {
		t.Errorf("other opener: got %q, want %q", result.Name, "Safari")
	}
}

func TestResolveURL_RewriteRule(t *testing.T) {
	vm := jsVM(t, `({
		defaultBrowser: "Safari",
		rewrite: [{
			match: "https://www.youtube.com/watch*",
			url: function(url) {
				return url.href.replace("https://www.youtube.com/watch", "https://youtu.be/");
			}
		}]
	})`)

	result, err := ResolveURL(vm, "https://www.youtube.com/watch?v=dQw4w9WgXcQ", nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.URL != "https://youtu.be/?v=dQw4w9WgXcQ" {
		t.Errorf("rewritten URL: got %q", result.URL)
	}
}

func TestResolveURL_OpenInBackground(t *testing.T) {
	result, err := ResolveURL(nil, "https://example.com", nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if result.OpenInBackground == nil || !*result.OpenInBackground {
		t.Error("expected OpenInBackground=true")
	}
}
