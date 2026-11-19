package main

import (
	"bytes"
	"encoding/base64"
	"finicky/browser"
	"finicky/config"
	"log/slog"
	"reflect"
	"strings"
	"testing"
)

func requestTestVM(t *testing.T, options string) *config.VM {
	t.Helper()
	requestVM, err := config.NewFromScript(finickyConfigAPIJS, "finickyConfig", `var finickyConfig = {
		defaultBrowser: { name: "Safari", args: ["--fixture"] }, `+options+`
	}`)
	if err != nil {
		t.Fatal(err)
	}
	return requestVM
}

func mainRequestLogs(t *testing.T) *bytes.Buffer {
	t.Helper()
	previousVM, previousDryRun, previousError := vm, dryRun, lastError
	previousLogger := slog.Default()
	t.Cleanup(func() {
		vm, dryRun, lastError = previousVM, previousDryRun, previousError
		slog.SetDefault(previousLogger)
	})
	dryRun = true
	var output bytes.Buffer
	slog.SetDefault(slog.New(slog.NewTextHandler(&output, &slog.HandlerOptions{Level: slog.LevelDebug})))
	return &output
}

func TestRequestInvocationSnapshots(t *testing.T) {
	quiet := requestTestVM(t, "options: { logRequests: false }")
	verbose := requestTestVM(t, "options: { logRequests: true }")
	omitted := requestTestVM(t, "")
	output := mainRequestLogs(t)
	var firstResult *browser.BrowserConfig
	for i, requestVM := range []*config.VM{quiet, verbose, quiet, omitted, nil} {
		vm = requestVM
		if requestVM != nil {
			_, err := requestVM.Runtime().RunString(`if (typeof optionReads === "undefined") {
				finickyConfigAPI = Object.assign({}, finickyConfigAPI, { getOption: (function(previousGetOption) {
					return function(name, config, fallback) {
						if (name === "logRequests") optionReads++;
						return previousGetOption(name, config, fallback);
					};
				})(finickyConfigAPI.getOption) });
			}
			var optionReads = 0;`)
			if err != nil {
				t.Fatal(err)
			}
		}
		output.Reset()
		result, err := resolveTestURL("https://example.com/path")
		if err != nil {
			t.Fatal(err)
		}
		if i == 0 {
			firstResult = result
		}
		if i < 4 && !reflect.DeepEqual(result, firstResult) {
			t.Fatalf("GUI route changed: %+v vs %+v", result, firstResult)
		}
		for _, marker := range []string{"Testing URL", "No opener detected", "Final browser options"} {
			if strings.Contains(output.String(), marker) != (i == 1) {
				t.Errorf("GUI invocation %d %s: %s", i, marker, output.String())
			}
		}
		output.Reset()
		openURL(URLInfo{URL: "https://example.com/path", OpenInBackground: true})
		for _, marker := range []string{"URL received", "Time taken evaluating URL and opening browser", "Starting browser", "Would run command (dry run)"} {
			if strings.Contains(output.String(), marker) != (i == 1) {
				t.Errorf("external invocation %d %s: %s", i, marker, output.String())
			}
		}
		if requestVM != nil && requestVM.Runtime().Get("optionReads").ToInteger() != 2 {
			t.Fatalf("expected one logRequests read per invocation, got %v", requestVM.Runtime().Get("optionReads"))
		}
	}
}

func TestProtocolRequestLogging(t *testing.T) {
	quiet := requestTestVM(t, "options: { logRequests: false }")
	verbose := requestTestVM(t, "options: { logRequests: true }")
	output := mainRequestLogs(t)
	original := "finicky://open/" + base64.StdEncoding.EncodeToString([]byte("https://example.com/path"))
	info := newURLInfo(original, nil, true)
	if info.URL != "https://example.com/path" || info.OriginalURL != original || !info.OpenInBackground {
		t.Fatalf("incorrect decode: %+v", info)
	}
	if output.Len() != 0 {
		t.Fatalf("successful decode logged in callback: %s", output.String())
	}
	for _, requestVM := range []*config.VM{quiet, verbose} {
		vm = requestVM
		output.Reset()
		openURL(info)
		logs := output.String()
		if strings.Contains(logs, "Decoded finicky protocol URL") != (requestVM == verbose) {
			t.Fatalf("decode logging: %s", logs)
		}
		if requestVM == verbose && strings.Index(logs, "Decoded finicky protocol URL") > strings.Index(logs, "URL received") {
			t.Fatalf("decode log order: %s", logs)
		}
	}
	output.Reset()
	vm = quiet
	bad := "finicky://open/%%%"
	info = newURLInfo(bad, nil, false)
	if info.URL != bad || info.OriginalURL != "" || !strings.Contains(output.String(), "level=WARN msg=\"Failed to decode finicky protocol URL\"") {
		t.Fatalf("decode failure changed: %+v %s", info, output.String())
	}
}

func TestGUIRequestRetainsEvaluationError(t *testing.T) {
	requestVM, err := config.NewFromScript(finickyConfigAPIJS, "finickyConfig", `var finickyConfig = {
		defaultBrowser: "Safari", options: { logRequests: false },
		handlers: [{ match: function() { throw new Error("fixture failure"); }, browser: "Firefox" }]
	}`)
	if err != nil {
		t.Fatal(err)
	}
	mainRequestLogs(t)
	vm = requestVM
	result, err := resolveTestURL("https://example.com/path")
	if err == nil || !strings.Contains(err.Error(), "fixture failure") || result == nil || result.Name != "com.apple.Safari" {
		t.Fatalf("GUI error/fallback changed: %+v %v", result, err)
	}
}
