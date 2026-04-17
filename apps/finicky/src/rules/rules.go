package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Rule struct {
	Match   []string `json:"match"`
	Browser string   `json:"browser"`
	Profile string   `json:"profile,omitempty"`
}

// UnmarshalJSON accepts both a single string and an array for the match field.
func (r *Rule) UnmarshalJSON(data []byte) error {
	var raw struct {
		Match   json.RawMessage `json:"match"`
		Browser string          `json:"browser"`
		Profile string          `json:"profile,omitempty"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.Browser = raw.Browser
	r.Profile = raw.Profile
	if raw.Match != nil {
		var s string
		if err := json.Unmarshal(raw.Match, &s); err == nil {
			r.Match = []string{s}
			return nil
		}
		return json.Unmarshal(raw.Match, &r.Match)
	}
	return nil
}

// MarshalJSON serializes match as a plain string when there is only one entry.
func (r Rule) MarshalJSON() ([]byte, error) {
	type RuleAlias struct {
		Match   interface{} `json:"match"`
		Browser string      `json:"browser"`
		Profile string      `json:"profile,omitempty"`
	}
	var match interface{}
	if len(r.Match) == 1 {
		match = r.Match[0]
	} else {
		match = r.Match
	}
	return json.Marshal(RuleAlias{Match: match, Browser: r.Browser, Profile: r.Profile})
}

type Options struct {
	KeepRunning     *bool `json:"keepRunning,omitempty"`
	HideIcon        *bool `json:"hideIcon,omitempty"`
	LogRequests     *bool `json:"logRequests,omitempty"`
	CheckForUpdates *bool `json:"checkForUpdates,omitempty"`
}

type RulesFile struct {
	DefaultBrowser string   `json:"defaultBrowser"`
	DefaultProfile string   `json:"defaultProfile,omitempty"`
	Options        *Options `json:"options,omitempty"`
	Rules          []Rule   `json:"rules"`
}

var customPath string

// SetCustomPath overrides the default rules.json location. Pass an empty
// string to revert to the default. Intended for testing and CLI flags.
func SetCustomPath(path string) {
	customPath = path
}

// GetPath returns the path to the rules JSON file.
// Returns the custom path if one was set via SetCustomPath, otherwise
// ~/Library/Application Support/Finicky/rules.json.
func GetPath() (string, error) {
	if customPath != "" {
		return customPath, nil
	}
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "Finicky", "rules.json"), nil
}

// Load reads the rules file from the default path. Returns an empty RulesFile if it doesn't exist.
func Load() (RulesFile, error) {
	path, err := GetPath()
	if err != nil {
		return RulesFile{}, err
	}
	return LoadFromPath(path)
}

// LoadFromPath reads a rules file from the given path. Returns an empty RulesFile if it doesn't exist.
func LoadFromPath(path string) (RulesFile, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return RulesFile{Rules: []Rule{}}, nil
	}
	if err != nil {
		return RulesFile{}, err
	}

	var rf RulesFile
	if err := json.Unmarshal(data, &rf); err != nil {
		return RulesFile{}, err
	}
	if rf.Rules == nil {
		rf.Rules = []Rule{}
	}
	return rf, nil
}

// Save writes the rules file to the default path, creating the directory if needed.
func Save(rf RulesFile) error {
	path, err := GetPath()
	if err != nil {
		return err
	}
	return SaveToPath(rf, path)
}

// SaveToPath writes the rules file to the given path, creating the directory if needed.
func SaveToPath(rf RulesFile, path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(rf, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// ToJSHandlers converts rules to the handler format expected by finickyConfigAPI.
// Rules with an empty match or browser are skipped.
func ToJSHandlers(rules []Rule) []map[string]interface{} {
	handlers := make([]map[string]interface{}, 0, len(rules))
	for _, r := range rules {
		// Filter out empty patterns
		matches := make([]string, 0, len(r.Match))
		for _, m := range r.Match {
			if m != "" {
				matches = append(matches, m)
			}
		}
		if len(matches) == 0 || r.Browser == "" {
			continue
		}
		var matchVal interface{}
		if len(matches) == 1 {
			matchVal = matches[0]
		} else {
			matchVal = matches
		}
		var browser interface{}
		if r.Profile != "" {
			browser = map[string]interface{}{"name": r.Browser, "profile": r.Profile}
		} else {
			browser = r.Browser
		}
		handlers = append(handlers, map[string]interface{}{
			"match":   matchVal,
			"browser": browser,
		})
	}
	return handlers
}

// ToJSConfigScript generates a JavaScript config assignment for the given namespace.
// It produces a valid finickyConfig object that can be evaluated in the JS VM.
func ToJSConfigScript(rf RulesFile, namespace string) (string, error) {
	defaultBrowser := rf.DefaultBrowser
	if defaultBrowser == "" {
		defaultBrowser = "com.apple.Safari"
	}

	var defaultBrowserObj interface{}
	if rf.DefaultProfile != "" {
		defaultBrowserObj = map[string]string{"name": defaultBrowser, "profile": rf.DefaultProfile}
	} else {
		defaultBrowserObj = defaultBrowser
	}

	defaultBrowserJSON, err := json.Marshal(defaultBrowserObj)
	if err != nil {
		return "", fmt.Errorf("failed to marshal defaultBrowser: %v", err)
	}

	handlersJSON, err := json.Marshal(ToJSHandlers(rf.Rules))
	if err != nil {
		return "", fmt.Errorf("failed to marshal handlers: %v", err)
	}

	if rf.Options == nil {
		return fmt.Sprintf("var %s = {default: {defaultBrowser: %s, handlers: %s}};",
			namespace, string(defaultBrowserJSON), string(handlersJSON)), nil
	}

	opts := make(map[string]interface{})
	if rf.Options.KeepRunning != nil {
		opts["keepRunning"] = *rf.Options.KeepRunning
	}
	if rf.Options.HideIcon != nil {
		opts["hideIcon"] = *rf.Options.HideIcon
	}
	if rf.Options.LogRequests != nil {
		opts["logRequests"] = *rf.Options.LogRequests
	}
	if rf.Options.CheckForUpdates != nil {
		opts["checkForUpdates"] = *rf.Options.CheckForUpdates
	}

	optsJSON, err := json.Marshal(opts)
	if err != nil {
		return "", fmt.Errorf("failed to marshal options: %v", err)
	}

	return fmt.Sprintf("var %s = {default: {defaultBrowser: %s, handlers: %s, options: %s}};",
		namespace, string(defaultBrowserJSON), string(handlersJSON), string(optsJSON)), nil
}
