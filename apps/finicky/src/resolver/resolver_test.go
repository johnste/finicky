package resolver_test

import (
	"bytes"
	"log/slog"
	"os"
	"reflect"
	"strings"
	"testing"

	"finicky/config"
	. "finicky/resolver"
	"finicky/rules"
)

func captureRequestLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	var output bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() { slog.SetDefault(previous) })
	return &output
}

func TestResolveURL_RequestLogging(t *testing.T) {
	for _, option := range []string{"false", "true", "omitted", "nil"} {
		t.Run(option, func(t *testing.T) {
			var requestVM *config.VM
			if option != "nil" {
				options := ""
				if option != "omitted" {
					options = "options: { logRequests: " + option + " },"
				}
				requestVM = jsVM(t, `({`+options+`
					defaultBrowser: "Safari",
					handlers: [{
						match: function(url, context) {
							console.log("user-log"); console.warn("user-warn"); console.error("user-error");
							return context.opener && context.opener.bundleId === "test.opener";
						},
						browser: { name: "Google Chrome", profile: "Work", args: ["--incognito"] }
					}]
				})`)
			}
			output := captureRequestLogs(t)
			for _, opener := range []*OpenerInfo{nil, {Name: "Test", BundleID: "test.opener", Path: "/test", WindowTitle: "Fixture"}} {
				for _, background := range []bool{false, true} {
					output.Reset()
					result, err := ResolveURL(requestVM, "https://example.com/path", opener, background, requestVM.GetAllConfigOptions().LogRequests)
					if err != nil {
						t.Fatal(err)
					}
					logs := output.String()
					wantRoutine := option == "true"
					for _, marker := range []string{"Final browser options", "No opener detected"} {
						want := wantRoutine && requestVM != nil && (marker != "No opener detected" || opener == nil)
						if strings.Contains(logs, marker) != want {
							t.Errorf("%s presence != %v: %s", marker, want, logs)
						}
					}
					if strings.Contains(logs, "Setting opener") != (wantRoutine && opener != nil) {
						t.Errorf("opener logging: %s", logs)
					}
					if requestVM != nil {
						for _, marker := range []string{"user-log", "user-warn", "user-error"} {
							if !strings.Contains(logs, marker) {
								t.Errorf("missing %s: %s", marker, logs)
							}
						}
					}
					other, otherErr := ResolveURL(requestVM, "https://example.com/path", opener, background, !wantRoutine)
					if otherErr != nil || !reflect.DeepEqual(result, other) {
						t.Fatalf("logging changed routing: %+v, %+v, %v", result, other, otherErr)
					}
					if result.OpenInBackground == nil || *result.OpenInBackground != background || result.URL != "https://example.com/path" {
						t.Fatalf("unexpected URL/background: %+v", result)
					}
					if requestVM != nil && opener != nil && (result.Name != "Google Chrome" || result.Profile != "Work" || !reflect.DeepEqual(result.Args, []string{"--incognito"})) {
						t.Fatalf("unexpected route: %+v", result)
					}
				}
			}
		})
	}
}

func TestResolveURL_RetainsFailureDiagnostics(t *testing.T) {
	requestVM := jsVM(t, `({ defaultBrowser: "Safari" })`)
	output := captureRequestLogs(t)
	for _, enabled := range []bool{false, true} {
		output.Reset()
		_, _ = ResolveURL(requestVM, "http://%", nil, false, enabled)
		if !strings.Contains(output.String(), "level=INFO msg=\"Failed to resolve short URL\"") {
			t.Fatalf("missing diagnostic: %s", output.String())
		}
	}
}

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
	result, err := ResolveURL(nil, "https://example.com", nil, false, false)
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
			result, err := ResolveURL(vm, tt.url, nil, false, false)
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
			{Match: []string{"*github.com/*"}, Browser: "Google Chrome"},
			{Match: []string{"https://linear.app/*"}, Browser: "Safari"},
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
			result, err := ResolveURL(vm, tt.url, nil, false, false)
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
			{Match: []string{"*github.com/*"}, Browser: "Google Chrome", Profile: "Work"},
		},
	}
	vm := rulesVM(t, rf)

	result, err := ResolveURL(vm, "https://github.com/foo", nil, false, false)
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
			result, err := ResolveURL(jsConfig, tt.url, nil, false, false)
			if err != nil {
				t.Fatal(err)
			}
			if result.Name != tt.browser {
				t.Errorf("got %q, want %q", result.Name, tt.browser)
			}
		})
	}
}

// TestResolveURL_MergedJSAndJSON_WithCachedRules verifies that JSON rules are
// applied when a JS config is present and cached rules have been seeded
// (simulating the fix for the startup bug where SetCachedRules was not called
// when a JS config existed).
func TestResolveURL_MergedJSAndJSON_WithCachedRules(t *testing.T) {
	jsConfig := jsVM(t, `({
		defaultBrowser: "Safari",
		handlers: [
			{ match: "*github.com/*", browser: "Firefox" }
		]
	})`)

	SetCachedRules(rules.RulesFile{
		DefaultBrowser: "Safari",
		Rules: []rules.Rule{
			{Match: []string{"linear.app/*"}, Browser: "Google Chrome"},
		},
	})
	t.Cleanup(func() { SetCachedRules(rules.RulesFile{}) })

	tests := []struct {
		url     string
		browser string
	}{
		{"https://github.com/foo", "Firefox"},                // JS handler wins
		{"https://linear.app/team/issue/1", "Google Chrome"}, // JSON rule applies
		{"https://example.com", "Safari"},                    // JS default
	}
	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result, err := ResolveURL(jsConfig, tt.url, nil, false, false)
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

	result, err := ResolveURL(vm, "https://example.com", terminal, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.Name != "Firefox" {
		t.Errorf("terminal opener: got %q, want %q", result.Name, "Firefox")
	}

	result, err = ResolveURL(vm, "https://example.com", other, false, false)
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

	result, err := ResolveURL(vm, "https://www.youtube.com/watch?v=dQw4w9WgXcQ", nil, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.URL != "https://youtu.be/?v=dQw4w9WgXcQ" {
		t.Errorf("rewritten URL: got %q", result.URL)
	}
}

func TestResolveURL_OpenInBackground(t *testing.T) {
	result, err := ResolveURL(nil, "https://example.com", nil, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if result.OpenInBackground == nil || !*result.OpenInBackground {
		t.Error("expected OpenInBackground=true")
	}
}
