package rules

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Rule struct {
	Match   string `json:"match"`
	Browser string `json:"browser"`
	Profile string `json:"profile,omitempty"`
}

type RulesFile struct {
	DefaultBrowser string `json:"defaultBrowser"`
	DefaultProfile string `json:"defaultProfile,omitempty"`
	Rules          []Rule `json:"rules"`
}

// GetPath returns the path to the rules JSON file:
// ~/Library/Application Support/Finicky/rules.json
func GetPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "Finicky", "rules.json"), nil
}

// Load reads the rules file from disk. Returns an empty RulesFile if it doesn't exist.
func Load() (RulesFile, error) {
	path, err := GetPath()
	if err != nil {
		return RulesFile{}, err
	}

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

// Save writes the rules file to disk, creating the directory if needed.
func Save(rf RulesFile) error {
	path, err := GetPath()
	if err != nil {
		return err
	}

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
		if r.Match == "" || r.Browser == "" {
			continue
		}
		var browser interface{}
		if r.Profile != "" {
			browser = map[string]interface{}{"name": r.Browser, "profile": r.Profile}
		} else {
			browser = r.Browser
		}
		handlers = append(handlers, map[string]interface{}{
			"match":   r.Match,
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

	defaultBrowserJSON, err := json.Marshal(defaultBrowser)
	if err != nil {
		return "", fmt.Errorf("failed to marshal defaultBrowser: %v", err)
	}

	handlersJSON, err := json.Marshal(ToJSHandlers(rf.Rules))
	if err != nil {
		return "", fmt.Errorf("failed to marshal handlers: %v", err)
	}

	return fmt.Sprintf("var %s = {defaultBrowser: %s, handlers: %s};",
		namespace, string(defaultBrowserJSON), string(handlersJSON)), nil
}
