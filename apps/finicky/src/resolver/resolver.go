package resolver

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"finicky/browser"
	"finicky/config"
	"finicky/rules"
	"finicky/shorturl"

	"github.com/dop251/goja"
)

// OpenerInfo describes the process that triggered the URL open.
type OpenerInfo struct {
	Name        string `json:"name"`
	BundleID    string `json:"bundleId"`
	Path        string `json:"path"`
	WindowTitle string `json:"windowTitle,omitempty"`
}

var (
	cachedRulesMu   sync.Mutex
	cachedRulesFile rules.RulesFile
)

// SetCachedRules stores a snapshot of the JSON rules so evaluateURL can use
// them without hitting disk on every URL open.
func SetCachedRules(rf rules.RulesFile) {
	cachedRulesMu.Lock()
	cachedRulesFile = rf
	cachedRulesMu.Unlock()
}

func getCachedRules() rules.RulesFile {
	cachedRulesMu.Lock()
	defer cachedRulesMu.Unlock()
	return cachedRulesFile
}

// ResolveURL determines which browser to use for the given URL.
//
// vm may be nil (no configuration at all). Whether to merge JSON rules is
// derived from vm.IsJSConfig().
//
// Always returns a non-nil config. Returns a non-nil error only when JS
// evaluation failed.
func ResolveURL(vm *config.VM, urlStr string, opener *OpenerInfo, openInBackground bool) (*browser.BrowserConfig, error) {
	if vm != nil {
		cfg, err := evaluateURL(vm, urlStr, opener)
		if err != nil {
			return defaultBrowserConfig(urlStr, openInBackground), err
		}
		cfg.OpenInBackground = mergeBackground(cfg.OpenInBackground, openInBackground)
		return cfg, nil
	}
	return defaultBrowserConfig(urlStr, openInBackground), nil
}

func mergeBackground(fromConfig *bool, requested bool) *bool {
	if fromConfig != nil {
		return fromConfig
	}
	return &requested
}

func evaluateURL(vm *config.VM, url string, opener *OpenerInfo) (*browser.BrowserConfig, error) {
	// Resolve short URLs before taking the VM lock, since it may hit the network.
	originalURL := url
	resolvedURL, err := shorturl.ResolveURL(url)
	if err != nil {
		slog.Info("Failed to resolve short URL", "error", err, "url", url, "using", resolvedURL)
	}

	var openerMap map[string]interface{}
	if opener != nil {
		openerMap = map[string]interface{}{
			"name":     opener.Name,
			"bundleId": opener.BundleID,
			"path":     opener.Path,
		}
		if opener.WindowTitle != "" {
			openerMap["windowTitle"] = opener.WindowTitle
		}
		slog.Debug("Setting opener", "name", opener.Name, "bundleId", opener.BundleID, "path", opener.Path, "windowTitle", opener.WindowTitle)
	} else {
		slog.Debug("No opener detected")
	}

	// When there is a JS config, append cached JSON rules as lower-priority handlers.
	var evalScript string
	var jsonHandlers interface{}
	if vm.IsJSConfig() {
		jsonHandlers = rules.ToJSHandlers(getCachedRules().Rules)
		evalScript = `finickyConfigAPI.openUrl(url, opener, originalUrl, Object.assign({}, finalConfig, {
			handlers: (finalConfig.handlers || []).concat(_jsonHandlers)
		}))`
	} else {
		evalScript = "finickyConfigAPI.openUrl(url, opener, originalUrl, finalConfig)"
	}

	// The globals below are shared by every evaluation, so setting them and
	// running the script must happen under one lock; otherwise a concurrent
	// evaluation (e.g. /api/test-url while a real URL is opening) could swap
	// in its own url or opener mid-run.
	var resultJSON interface{}
	err = vm.Run(func(runtime *goja.Runtime) error {
		runtime.Set("originalUrl", originalURL)
		runtime.Set("url", resolvedURL)
		if openerMap != nil {
			runtime.Set("opener", openerMap)
		} else {
			runtime.Set("opener", nil)
		}
		if jsonHandlers != nil {
			runtime.Set("_jsonHandlers", jsonHandlers)
		}
		openResult, err := runtime.RunString(evalScript)
		if err != nil {
			return fmt.Errorf("failed to evaluate URL in config: %v", err)
		}
		resultJSON = openResult.ToObject(runtime).Export()
		return nil
	})
	if err != nil {
		return nil, err
	}

	resultBytes, err := json.Marshal(resultJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to process browser configuration: %v", err)
	}

	var browserResult browser.BrowserResult
	if err := json.Unmarshal(resultBytes, &browserResult); err != nil {
		return nil, fmt.Errorf("failed to parse browser configuration: %v", err)
	}

	slog.Debug("Final browser options",
		"name", browserResult.Browser.Name,
		"openInBackground", browserResult.Browser.OpenInBackground,
		"profile", browserResult.Browser.Profile,
		"args", browserResult.Browser.Args,
		"appType", browserResult.Browser.AppType,
	)

	var resultErr error
	if browserResult.Error != "" {
		resultErr = fmt.Errorf("%s", browserResult.Error)
	}
	return &browserResult.Browser, resultErr
}

func defaultBrowserConfig(urlStr string, openInBackground bool) *browser.BrowserConfig {
	bg := openInBackground
	return &browser.BrowserConfig{
		Name:             "com.apple.Safari",
		AppType:          "bundleId",
		OpenInBackground: &bg,
		Args:             []string{},
		URL:              urlStr,
	}
}
