package rules_test

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	. "finicky/rules"
)

// ---- ToJSHandlers ----

func TestToJSHandlers_Empty(t *testing.T) {
	result := ToJSHandlers([]Rule{})
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %d entries", len(result))
	}
}

func TestToJSHandlers_SkipsIncompleteRules(t *testing.T) {
	rules := []Rule{
		{Match: []string{""}, Browser: "Firefox"},         // no match
		{Match: []string{"example.com"}, Browser: ""},     // no browser
		{Match: []string{"example.com"}, Browser: "Safari"}, // valid
	}
	result := ToJSHandlers(rules)
	if len(result) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(result))
	}
}

func TestToJSHandlers_StringBrowser(t *testing.T) {
	rules := []Rule{
		{Match: []string{"*github.com/*"}, Browser: "Firefox"},
	}
	result := ToJSHandlers(rules)
	if len(result) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(result))
	}
	h := result[0]
	if h["match"] != "*github.com/*" {
		t.Errorf("match: got %q, want %q", h["match"], "*github.com/*")
	}
	if h["browser"] != "Firefox" {
		t.Errorf("browser: got %v, want %q", h["browser"], "Firefox")
	}
}

func TestToJSHandlers_WithProfile(t *testing.T) {
	rules := []Rule{
		{Match: []string{"*github.com/*"}, Browser: "Google Chrome", Profile: "Work"},
	}
	result := ToJSHandlers(rules)
	if len(result) != 1 {
		t.Fatalf("expected 1 handler, got %d", len(result))
	}
	browser, ok := result[0]["browser"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected browser to be a map, got %T", result[0]["browser"])
	}
	if browser["name"] != "Google Chrome" {
		t.Errorf("name: got %q, want %q", browser["name"], "Google Chrome")
	}
	if browser["profile"] != "Work" {
		t.Errorf("profile: got %q, want %q", browser["profile"], "Work")
	}
}

func TestToJSHandlers_MultipleRules(t *testing.T) {
	rules := []Rule{
		{Match: []string{"*github.com/*"}, Browser: "Firefox"},
		{Match: []string{"https://linear.app/*"}, Browser: "Google Chrome", Profile: "Work"},
		{Match: []string{"example.com"}, Browser: "Safari"},
	}
	result := ToJSHandlers(rules)
	if len(result) != 3 {
		t.Fatalf("expected 3 handlers, got %d", len(result))
	}
	// Order must be preserved
	if result[0]["match"] != "*github.com/*" {
		t.Errorf("wrong order: first match got %q", result[0]["match"])
	}
}

// ---- ToJSConfigScript ----

func TestToJSConfigScript_DefaultBrowserFallback(t *testing.T) {
	rf := RulesFile{Rules: []Rule{{Match: []string{"example.com"}, Browser: "Firefox"}}}
	script, err := ToJSConfigScript(rf, "finickyConfig")
	if err != nil {
		t.Fatal(err)
	}
	// Should fall back to com.apple.Safari when no defaultBrowser is set
	if script == "" {
		t.Error("expected non-empty script")
	}
	// com.apple.Safari should appear as the default
	if !contains(script, "com.apple.Safari") {
		t.Errorf("expected fallback default browser in script, got: %s", script)
	}
}

func TestToJSConfigScript_ExplicitDefaultBrowser(t *testing.T) {
	rf := RulesFile{DefaultBrowser: "Firefox", Rules: []Rule{}}
	script, err := ToJSConfigScript(rf, "finickyConfig")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(script, "Firefox") {
		t.Errorf("expected Firefox in script, got: %s", script)
	}
}

func TestToJSConfigScript_NamespaceIsUsed(t *testing.T) {
	rf := RulesFile{DefaultBrowser: "Safari"}
	script, err := ToJSConfigScript(rf, "myNamespace")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(script, "myNamespace") {
		t.Errorf("expected namespace in script, got: %s", script)
	}
}

// ---- Load / Save round-trip ----

func TestLoadSave_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")

	original := RulesFile{
		DefaultBrowser: "Firefox",
		DefaultProfile: "Work",
		Rules: []Rule{
			{Match: []string{"*github.com/*"}, Browser: "Google Chrome", Profile: "Personal"},
			{Match: []string{"https://linear.app/*"}, Browser: "Safari"},
		},
	}

	if err := SaveToPath(original, path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := LoadFromPath(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if loaded.DefaultBrowser != original.DefaultBrowser {
		t.Errorf("DefaultBrowser: got %q, want %q", loaded.DefaultBrowser, original.DefaultBrowser)
	}
	if loaded.DefaultProfile != original.DefaultProfile {
		t.Errorf("DefaultProfile: got %q, want %q", loaded.DefaultProfile, original.DefaultProfile)
	}
	if len(loaded.Rules) != len(original.Rules) {
		t.Fatalf("Rules length: got %d, want %d", len(loaded.Rules), len(original.Rules))
	}
	for i, r := range original.Rules {
		got := loaded.Rules[i]
		if !reflect.DeepEqual(got.Match, r.Match) || got.Browser != r.Browser || got.Profile != r.Profile {
			t.Errorf("Rule[%d]: got %+v, want %+v", i, got, r)
		}
	}
}

func TestLoad_MissingFile(t *testing.T) {
	_, err := LoadFromPath(filepath.Join(t.TempDir(), "nonexistent.json"))
	if err != nil {
		t.Errorf("expected nil error for missing file, got %v", err)
	}
}

func TestLoad_EmptyRulesNotNil(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "rules.json")
	if err := os.WriteFile(path, []byte(`{"defaultBrowser":"Safari"}`), 0644); err != nil {
		t.Fatal(err)
	}
	rf, err := LoadFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if rf.Rules == nil {
		t.Error("expected Rules to be non-nil slice, got nil")
	}
}

func contains(s, substr string) bool {
	return strings.Contains(s, substr)
}
