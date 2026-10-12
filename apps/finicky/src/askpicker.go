package main

import (
	"fmt"
	"log/slog"

	"finicky/browser"
	"finicky/picker"
)

// resolveAskConfig presents a picker for a browser config with appType "ask"
// and returns the destination the user chose. The second return value is false
// when nothing should be opened, either because the user cancelled or because
// the config offered no options to choose from.
func resolveAskConfig(config browser.BrowserConfig) (browser.BrowserConfig, bool) {
	if len(config.Options) == 0 {
		slog.Error("Browser appType is \"ask\" but no options were configured", "url", config.URL)
		return config, false
	}

	labels := make([]string, 0, len(config.Options))
	for _, option := range config.Options {
		labels = append(labels, askOptionLabel(option))
	}

	title := config.Title
	if title == "" {
		title = "Open link in"
	}

	choice := picker.Choose(title, config.URL, labels)
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
