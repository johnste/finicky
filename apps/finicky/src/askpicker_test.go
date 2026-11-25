package main

import (
	"testing"

	"finicky/browser"
)

func TestAskOptionLabel(t *testing.T) {
	tests := []struct {
		name   string
		option browser.BrowserConfig
		want   string
	}{
		{
			name:   "explicit title wins",
			option: browser.BrowserConfig{Name: "Google Chrome", Profile: "Profile 2", Title: "Work"},
			want:   "Work",
		},
		{
			name:   "profile qualifies the name",
			option: browser.BrowserConfig{Name: "Google Chrome", Profile: "Profile 2"},
			want:   "Google Chrome - Profile 2",
		},
		{
			name:   "bare name when there is no profile",
			option: browser.BrowserConfig{Name: "Safari"},
			want:   "Safari",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := askOptionLabel(test.option); got != test.want {
				t.Errorf("askOptionLabel() = %q, want %q", got, test.want)
			}
		})
	}
}

// An "ask" config with nothing to choose from must not open anything, and must
// not reach the picker, which would block on a modal nobody asked for.
func TestResolveAskConfigWithoutOptions(t *testing.T) {
	config := browser.BrowserConfig{
		Name:    "ask",
		AppType: "ask",
		URL:     "https://example.com",
	}

	got, shouldLaunch := resolveAskConfig(config, false)

	if shouldLaunch {
		t.Error("expected shouldLaunch to be false when no options are configured")
	}
	if got.URL != config.URL {
		t.Errorf("expected the url to be preserved, got %q", got.URL)
	}
}

func TestAbbreviateURL(t *testing.T) {
	tests := []struct {
		name      string
		url       string
		maxLength int
		want      string
	}{
		{
			name:      "short urls are untouched",
			url:       "https://example.com/docs",
			maxLength: 90,
			want:      "https://example.com/docs",
		},
		{
			name:      "a url exactly at the cap is untouched",
			url:       "https://example.com/12345",
			maxLength: 25,
			want:      "https://example.com/12345",
		},
		{
			name:      "the middle goes, scheme and host survive",
			url:       "https://mail.google.com/mail/u/0/some/very/long/path/that/keeps/going#inbox/FMfcgzQbfWKpXsMb",
			maxLength: 40,
			want:      "https://mail.google.com…FMfcgzQbfWKpXsMb",
		},
		{
			name:      "a host longer than the cap falls back to the tail",
			url:       "https://a-really-long-hostname-beyond-the-budget.example.com/path",
			maxLength: 20,
			want:      "https://a-really-lo…",
		},
		{
			name:      "a url with no path keeps the whole host",
			url:       "https://averyveryverylonghostnamethatfitsthebudget.example.com",
			maxLength: 40,
			want:      "https://averyveryverylonghostnamethatfi…",
		},
		{
			name:      "a schemeless string still truncates",
			url:       "not-really-a-url-but-still-far-too-long-to-show",
			maxLength: 20,
			want:      "not-really-a-url-bu…",
		},
		{
			name:      "multibyte characters are not split",
			url:       "https://example.com/äöüäöüäöüäöüäöüäöü",
			maxLength: 26,
			want:      "https://example.com…äöüäöü",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := abbreviateURL(test.url, test.maxLength)
			if got != test.want {
				t.Errorf("abbreviateURL() = %q, want %q", got, test.want)
			}
			if length := len([]rune(got)); length > test.maxLength {
				t.Errorf("result is %d runes, over the %d cap", length, test.maxLength)
			}
		})
	}
}

// A dry run must not reach the picker. If this regresses the test blocks on a
// modal rather than failing an assertion, which is itself the signal.
func TestResolveAskConfigDryRunDoesNotPrompt(t *testing.T) {
	config := browser.BrowserConfig{
		Name:    "ask",
		AppType: "ask",
		URL:     "https://example.com",
		Options: []browser.BrowserConfig{
			{Name: "Google Chrome", Profile: "Profile 2", Title: "Work"},
			{Name: "Google Chrome", Profile: "Default", Title: "Personal"},
		},
	}

	if _, shouldLaunch := resolveAskConfig(config, true); shouldLaunch {
		t.Error("a dry run must not resolve to something launchable")
	}
}
