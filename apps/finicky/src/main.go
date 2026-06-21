package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreServices
#include <stdlib.h>
#include "main.h"
*/
import "C"

import (
	_ "embed"
	"encoding/base64"
	"finicky/browser"
	"finicky/config"
	"finicky/logger"
	"finicky/resolver"
	"finicky/rules"
	"finicky/util"
	"finicky/version"
	"finicky/window"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
)

//go:embed assets/finickyConfigAPI.js
var finickyConfigAPIJS []byte

type UpdateInfo struct {
	ReleaseInfo        *version.ReleaseInfo
	UpdateCheckEnabled bool
}

type URLInfo struct {
	URL              string
	Opener           *resolver.OpenerInfo
	OpenInBackground bool
}

type ConfigInfo struct {
	Handlers       int16  `json:"handlers"`
	Rewrites       int16  `json:"rewrites"`
	DefaultBrowser string `json:"defaultBrowser"`
	ConfigPath     string `json:"configPath"`
}

var urlListener chan URLInfo = make(chan URLInfo)
var windowClosed chan struct{} = make(chan struct{})
var vm *config.VM

var forceWindowOpen bool = false
var queueWindowOpen chan bool = make(chan bool)
var lastError error
var dryRun bool = false
var skipJSConfig bool = false
var updateInfo UpdateInfo
var configInfo *ConfigInfo
var lastConfigPayload map[string]interface{}
var shouldKeepRunning bool = true

// stateMu guards vm, updateInfo, configInfo, lastConfigPayload, and
// shouldKeepRunning. They're written from the single event-loop goroutine in
// main() (and from setupVM) but now also read/written from the REST API's
// HTTP handler goroutines (window.TestURLFunc, window.SaveRulesHandler,
// window.GetConfigFunc, window.GetUpdateInfoFunc), so plain reads/writes are
// no longer safe. Always go through the getX/setX helpers below instead of
// touching these globals directly.
var stateMu sync.Mutex

func getVM() *config.VM {
	stateMu.Lock()
	defer stateMu.Unlock()
	return vm
}

func setVM(v *config.VM) {
	stateMu.Lock()
	vm = v
	stateMu.Unlock()
}

func getUpdateInfo() UpdateInfo {
	stateMu.Lock()
	defer stateMu.Unlock()
	return updateInfo
}

func setUpdateInfo(ui UpdateInfo) {
	stateMu.Lock()
	updateInfo = ui
	stateMu.Unlock()
}

func getConfigInfo() *ConfigInfo {
	stateMu.Lock()
	defer stateMu.Unlock()
	return configInfo
}

func getConfigPayload() map[string]interface{} {
	stateMu.Lock()
	defer stateMu.Unlock()
	return lastConfigPayload
}

// publishConfigInfo records a freshly computed ConfigInfo and returns the
// resulting effective value. If ci is nil the previously published
// ConfigInfo is kept and returned (mirrors the pre-existing behavior in
// setupVM, where a nil ConfigState from the VM doesn't clear out the prior
// config).
func publishConfigInfo(ci *ConfigInfo) *ConfigInfo {
	stateMu.Lock()
	defer stateMu.Unlock()
	if ci != nil {
		configInfo = ci
	}
	return configInfo
}

func setConfigPayload(payload map[string]interface{}) {
	stateMu.Lock()
	lastConfigPayload = payload
	stateMu.Unlock()
}

func getShouldKeepRunning() bool {
	stateMu.Lock()
	defer stateMu.Unlock()
	return shouldKeepRunning
}

func setShouldKeepRunning(v bool) {
	stateMu.Lock()
	shouldKeepRunning = v
	stateMu.Unlock()
}

func main() {
	startTime := time.Now()
	logger.Setup()
	runtime.LockOSThread()

	// Define command line flags
	configPathPtr := flag.String("config", "", "Path to custom JS configuration file")
	rulesPathPtr := flag.String("rules", "", "Path to custom rules JSON file")
	noConfigPtr := flag.Bool("no-config", false, "Skip JS configuration file entirely")
	windowPtr := flag.Bool("window", false, "Force window to open")
	dryRunPtr := flag.Bool("dry-run", false, "Simulate without actually opening browsers")
	flag.Parse()

	// Use the parsed values
	customConfigPath := *configPathPtr
	if customConfigPath != "" {
		slog.Debug("Using custom config path", "path", customConfigPath)
	}

	if *rulesPathPtr != "" {
		slog.Debug("Using custom rules path", "path", *rulesPathPtr)
		rules.SetCustomPath(*rulesPathPtr)
	}

	if *noConfigPtr {
		slog.Debug("Skipping JS config")
		skipJSConfig = true
	}

	if *windowPtr {
		forceWindowOpen = true
	}

	dryRun = *dryRunPtr

	currentVersion := version.GetCurrentVersion()
	commitHash, buildDate := version.GetBuildInfo()
	slog.Info("Starting Finicky", "version", currentVersion)
	slog.Debug("Build info", "buildDate", buildDate, "commitHash", commitHash)

	go func() {
		is_default_browser, err := setDefaultBrowser()
		if err != nil {
			slog.Debug("Failed checking if we are the default browser", "error", err)
		} else if !is_default_browser {
			slog.Debug("Finicky is not the default browser")
		} else {
			slog.Debug("Finicky is the default browser")
		}
	}()

	namespace := "finickyConfig"
	configChange := make(chan struct{}, 1)
	cfw, err := config.NewConfigFileWatcher(customConfigPath, namespace, configChange)

	if err != nil {
		handleFatalError(fmt.Sprintf("Failed to setup config file watcher: %v", err))
	}

	initialVM, err := setupVM(cfw, namespace)
	if err != nil {
		handleFatalError(err.Error())
	}
	setVM(initialVM)

	slog.Debug("VM setup complete", "duration", fmt.Sprintf("%.2fms", float64(time.Since(startTime).Microseconds())/1000))

	go checkForUpdates()

	window.TestURLFunc = func(url string) (interface{}, error) {
		slog.Debug("Testing URL", "url", url)
		cfg, err := resolver.ResolveURL(getVM(), url, nil, false)
		if err != nil {
			return nil, err
		}
		return map[string]interface{}{
			"url":              cfg.URL,
			"browser":          cfg.Name,
			"openInBackground": cfg.OpenInBackground,
			"profile":          cfg.Profile,
			"args":             cfg.Args,
		}, nil
	}

	window.GetVersionFunc = version.GetCurrentVersion

	window.GetConfigFunc = func() interface{} {
		return getConfigPayload()
	}

	window.GetUpdateInfoFunc = func() interface{} {
		ui := getUpdateInfo()
		if ui.ReleaseInfo == nil && !ui.UpdateCheckEnabled {
			return nil
		}
		return buildUpdateInfoPayload(ui)
	}

	if err := window.StartAPIServer(); err != nil {
		handleFatalError(fmt.Sprintf("Failed to start API server: %v", err))
	}

	// Set up rules save handler.
	// When there is no JS config, rebuild the VM from the updated rules.
	// When there is a JS config, JSON rules are loaded fresh in evaluateURL — nothing to do.
	// Invoked synchronously from the /api/rules HTTP handler so the new VM is
	// guaranteed to be in place by the time that request completes.
	window.SaveRulesHandler = func(rf rules.RulesFile) {
		slog.Debug("Rules updated", "count", len(rf.Rules))
		resolver.SetCachedRules(rf)
		if v := getVM(); v == nil || !v.IsJSConfig() {
			if rf.DefaultBrowser == "" && len(rf.Rules) == 0 && rf.Options == nil {
				setVM(nil)
				return
			}
			script, err := rules.ToJSConfigScript(rf, namespace)
			if err != nil {
				slog.Error("Failed to generate config from rules", "error", err)
				return
			}
			newVM, err := config.NewFromScript(finickyConfigAPIJS, namespace, script)
			if err != nil {
				slog.Error("Failed to rebuild VM from rules", "error", err)
				return
			}
			setVM(newVM)
			if newVM != nil {
				setShouldKeepRunning(newVM.GetAllConfigOptions().KeepRunning)
				go checkForUpdates()
			}
		}
	}

	const oneDay = 24 * time.Hour

	var showingWindow bool = false
	timeoutChan := time.After(1 * time.Second)
	updateChan := time.After(oneDay)

	if v := getVM(); v != nil {
		setShouldKeepRunning(v.GetAllConfigOptions().KeepRunning)
	}
	if getShouldKeepRunning() {
		timeoutChan = nil
	}

	go func() {
		slog.Info("Listening for events...")
		for {
			select {
			case urlInfo := <-urlListener:
				startTime := time.Now()

				url := urlInfo.URL

				slog.Info("URL received", "url", url)

				config, err := resolver.ResolveURL(getVM(), url, urlInfo.Opener, urlInfo.OpenInBackground)
				if err != nil {
					handleRuntimeError(err)
				} else {
					lastError = nil
				}
				if launchErr := browser.LaunchBrowser(*config, dryRun, urlInfo.OpenInBackground); launchErr != nil {
					slog.Error("Failed to start browser", "error", launchErr)
				}

				slog.Debug("Time taken evaluating URL and opening browser", "duration", fmt.Sprintf("%.2fms", float64(time.Since(startTime).Microseconds())/1000))

				if !showingWindow && !getShouldKeepRunning() {
					timeoutChan = time.After(2 * time.Second)
				} else {
					timeoutChan = nil
				}

			case <-configChange:
				startTime := time.Now()
				slog.Debug("Config has changed")
				newVM, setupErr := setupVM(cfw, namespace)
				setVM(newVM)
				if setupErr != nil {
					handleRuntimeError(setupErr)
				} else {
					lastError = nil
					C.SetStatusItemError(false)
				}
				slog.Debug("VM refresh complete", "duration", fmt.Sprintf("%.2fms", float64(time.Since(startTime).Microseconds())/1000))
				if newVM != nil {
					setShouldKeepRunning(newVM.GetAllConfigOptions().KeepRunning)
					go checkForUpdates()
				}

			case shouldShowWindow := <-queueWindowOpen:
				if !showingWindow && shouldShowWindow {
					go ShowConfigWindow()
					showingWindow = true
					timeoutChan = nil
				}

			case <-updateChan:
				go checkForUpdates()
				updateChan = time.After(oneDay)

			case <-windowClosed:
				if !getShouldKeepRunning() {
					slog.Info("Exiting due to window closed")
					tearDown()
				} else {
					slog.Debug("Window closed")
				}

			case <-timeoutChan:
				slog.Info("Exiting due to timeout")
				tearDown()
			}
		}
	}()

	shouldHideIcon := false
	if v := getVM(); v != nil {
		shouldHideIcon = v.GetAllConfigOptions().HideIcon
	}
	C.RunApp(C.bool(forceWindowOpen), C.bool(!shouldHideIcon), C.bool(getShouldKeepRunning()))
}

func handleRuntimeError(err error) {
	slog.Error("Failed evaluating url", "error", err)
	lastError = err
	C.SetStatusItemError(true)
}

//export HandleURL
func HandleURL(url *C.char, name *C.char, bundleId *C.char, path *C.char, windowTitle *C.char, openInBackground C.bool) {
	var opener resolver.OpenerInfo

	if name != nil && bundleId != nil && path != nil {
		opener = resolver.OpenerInfo{
			Name:     C.GoString(name),
			BundleID: C.GoString(bundleId),
			Path:     C.GoString(path),
		}
		if windowTitle != nil {
			opener.WindowTitle = C.GoString(windowTitle)
		}
	}

	urlString := C.GoString(url)

	// Handle finicky:// protocol URLs by extracting and decoding the embedded URL
	if strings.HasPrefix(urlString, "finicky://open/") {
		encodedURL := strings.TrimPrefix(urlString, "finicky://open/")
		if decodedBytes, err := base64.StdEncoding.DecodeString(encodedURL); err == nil {
			urlString = string(decodedBytes)
			slog.Debug("Decoded finicky protocol URL", "original", C.GoString(url), "decoded", urlString)
		} else {
			slog.Warn("Failed to decode finicky protocol URL", "error", err, "url", C.GoString(url))
		}
	}

	urlListener <- URLInfo{
		URL:              urlString,
		Opener:           &opener,
		OpenInBackground: bool(openInBackground),
	}
}


func handleFatalError(errorMessage string) {
	slog.Error("Fatal error", "msg", errorMessage)
	lastError = fmt.Errorf("%s", errorMessage)
	C.SetStatusItemError(true)
}

//export QueueWindowDisplay
func QueueWindowDisplay(openWindow int32) {
	queueWindowOpen <- openWindow != 0
}

//export ShowConfigWindow
func ShowConfigWindow() {
	slog.Debug("Showing window")
	window.ShowWindow()
}

//export WindowDidClose
func WindowDidClose() {
	windowClosed <- struct{}{}
}

//export GetCurrentConfigPath
func GetCurrentConfigPath() *C.char {
	if ci := getConfigInfo(); ci != nil && ci.ConfigPath != "" {
		cPath := C.CString(ci.ConfigPath)
		return cPath
	} else {
		return nil
	}
}

func checkForUpdates() {
	var runtime *goja.Runtime
	if v := getVM(); v != nil {
		runtime = v.Runtime()
	}

	releaseInfo, updateCheckEnabled, err := version.CheckForUpdatesIfEnabled(runtime)
	if err != nil {
		slog.Error("Error checking for updates", "error", err)
	}

	ui := UpdateInfo{
		ReleaseInfo:        releaseInfo,
		UpdateCheckEnabled: updateCheckEnabled,
	}
	setUpdateInfo(ui)

	if ui.ReleaseInfo != nil && ui.ReleaseInfo.HasUpdate {
		slog.Info("New version is available", "version", ui.ReleaseInfo.LatestVersion)
	}

	window.BroadcastSSE("updateInfo", buildUpdateInfoPayload(ui))
}

func buildUpdateInfoPayload(ui UpdateInfo) map[string]interface{} {
	if ui.ReleaseInfo != nil {
		return map[string]interface{}{
			"version":            ui.ReleaseInfo.LatestVersion,
			"hasUpdate":          ui.ReleaseInfo.HasUpdate,
			"updateCheckEnabled": ui.UpdateCheckEnabled,
			"downloadUrl":        ui.ReleaseInfo.DownloadUrl,
			"releaseUrl":         ui.ReleaseInfo.ReleaseUrl,
		}
	}
	return map[string]interface{}{
		"version": "", "hasUpdate": false,
		"updateCheckEnabled": ui.UpdateCheckEnabled,
		"downloadUrl": "", "releaseUrl": "",
	}
}

func tearDown() {
	checkForUpdates()
	slog.Info("Exiting...")
	os.Exit(0)
}

func setupVM(cfw *config.ConfigFileWatcher, namespace string) (*config.VM, error) {
	logRequests := true
	var err error

	defer func() {
		err = logger.SetupFile(logRequests)
		if err != nil {
			slog.Warn("Failed to setup file logging", "error", err)
		}
	}()

	var currentBundlePath, configPath string
	if !skipJSConfig {
		var err2 error
		currentBundlePath, configPath, err2 = cfw.BundleConfig()
		if err2 != nil {
			return nil, fmt.Errorf("failed to read config: %v", err2)
		}
	}

	// Always seed the cached rules from disk so JSON rules are applied
	// immediately on startup, even when a JS config is also present.
	if rf, rulesErr := rules.Load(); rulesErr != nil {
		slog.Warn("Failed to pre-load rules cache", "error", rulesErr)
	} else {
		resolver.SetCachedRules(rf)
	}

	var newVM *config.VM

	if currentBundlePath != "" {
		newVM, err = config.New(finickyConfigAPIJS, namespace, currentBundlePath)
		if err != nil {
			return nil, fmt.Errorf("failed to setup VM: %v", err)
		}
	} else {
		rf, rulesErr := rules.Load()
		if rulesErr != nil {
			slog.Warn("Failed to load rules file", "error", rulesErr)
		} else {
			resolver.SetCachedRules(rf)
			if rf.DefaultBrowser != "" || len(rf.Rules) > 0 {
				script, scriptErr := rules.ToJSConfigScript(rf, namespace)
				if scriptErr != nil {
					return nil, fmt.Errorf("failed to generate config from rules: %v", scriptErr)
				}
				newVM, err = config.NewFromScript(finickyConfigAPIJS, namespace, script)
				if err != nil {
					return nil, fmt.Errorf("failed to setup VM from rules: %v", err)
				}
				configPath, _ = rules.GetPath()
			}
		}
	}

	if newVM == nil {
		return nil, nil
	}

	cs := newVM.GetConfigState()
	var ci *ConfigInfo
	if cs != nil {
		ci = &ConfigInfo{
			Handlers:       cs.Handlers,
			Rewrites:       cs.Rewrites,
			DefaultBrowser: cs.DefaultBrowser,
			ConfigPath:     configPath,
		}
	}

	opts := newVM.GetAllConfigOptions()
	logRequests = opts.LogRequests

	// publishConfigInfo keeps the previously published ConfigInfo when cs
	// (and so ci) is nil, mirroring the prior behavior of this function.
	publishedCI := publishConfigInfo(ci)
	payload := map[string]interface{}{
		"handlers":       publishedCI.Handlers,
		"rewrites":       publishedCI.Rewrites,
		"defaultBrowser": publishedCI.DefaultBrowser,
		"configPath":     util.ShortenPath(publishedCI.ConfigPath),
		"hasJsConfig":    newVM.IsJSConfig(),
		"options": map[string]interface{}{
			"keepRunning":     opts.KeepRunning,
			"hideIcon":        opts.HideIcon,
			"logRequests":     opts.LogRequests,
			"checkForUpdates": opts.CheckForUpdates,
		},
	}
	setConfigPayload(payload)
	window.BroadcastSSE("config", payload)

	return newVM, nil
}
