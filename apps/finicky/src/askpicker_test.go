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

	got, shouldLaunch := resolveAskConfig(config)

	if shouldLaunch {
		t.Error("expected shouldLaunch to be false when no options are configured")
	}
	if got.URL != config.URL {
		t.Errorf("expected the url to be preserved, got %q", got.URL)
	}
}
