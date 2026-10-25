package main

import (
	"fmt"
	"log/slog"
	"strings"
	"unicode/utf8"

	"finicky/browser"
	"finicky/picker"
)

// maxPickerURLLength caps the url shown in the picker. A url is one unbroken
// token, so AppKit cannot word-wrap it: shown whole, a long one widens the
// alert until its buttons are off screen. 60 keeps the dialog to about two
// lines; the untruncated url is still available as the tooltip.
const maxPickerURLLength = 60

// abbreviateURL shortens a url for display by eliding its middle, keeping the
// scheme and host whole because that is what the choice is actually made on,
// plus the tail, which is where two otherwise identical urls differ.
func abbreviateURL(rawURL string, maxLength int) string {
	runes := []rune(rawURL)
	if maxLength <= 1 || len(runes) <= maxLength {
		return rawURL
	}

	hostEnd := len(runes)
	if scheme := strings.Index(rawURL, "://"); scheme >= 0 {
		afterScheme := scheme + len("://")
		if slash := strings.IndexByte(rawURL[afterScheme:], '/'); slash >= 0 {
			hostEnd = utf8.RuneCountInString(rawURL[:afterScheme+slash])
		}
	}

	// A host that fills the budget on its own leaves no room to keep it whole,
	// so fall back to truncating the tail.
	if hostEnd+1 >= maxLength {
		return string(runes[:maxLength-1]) + "…"
	}

	tail := maxLength - hostEnd - 1
	return string(runes[:hostEnd]) + "…" + string(runes[len(runes)-tail:])
}

// resolveAskConfig presents a picker for a browser config with appType "ask"
// and returns the destination the user chose. The second return value is false
// when nothing should be opened: the user cancelled, the config offered no
// options to choose from, or this is a dry run.
func resolveAskConfig(config browser.BrowserConfig, dryRun bool) (browser.BrowserConfig, bool) {
	if len(config.Options) == 0 {
		slog.Error("Browser appType is \"ask\" but no options were configured", "url", config.URL)
		return config, false
	}

	labels := make([]string, 0, len(config.Options))
	for _, option := range config.Options {
		labels = append(labels, askOptionLabel(option))
	}

	// A dry run simulates without opening anything, and a modal waiting on a
	// click is the one thing a simulation must not do.
	if dryRun {
		slog.Info("Dry run, not showing the picker", "url", config.URL, "options", strings.Join(labels, ", "))
		return config, false
	}

	title := config.Title
	if title == "" {
		title = "Open link in"
	}

	// The url is abbreviated for display and passed whole as the tooltip.
	choice := picker.Choose(title, abbreviateURL(config.URL, maxPickerURLLength), config.URL, labels)
	if choice == picker.Cancelled {
		slog.Info("Picker cancelled, not opening url", "url", config.URL)
		return config, false
	}

	chosen := config.Options[choice]
	chosen.URL = config.URL
	if chosen.OpenInBackground == nil {
		chosen.OpenInBackground = config.OpenInBackground
	}

	slog.Info("Picker selected browser", "name", chosen.Name, "profile", chosen.Profile, "url", chosen.URL)
	return chosen, true
}

// askOptionLabel is the text shown on an option's button: an explicit title if
// the config gave one, otherwise the browser name qualified by its profile.
func askOptionLabel(option browser.BrowserConfig) string {
	if option.Title != "" {
		return option.Title
	}
	if option.Profile != "" {
		return fmt.Sprintf("%s - %s", option.Name, option.Profile)
	}
	return option.Name
}
