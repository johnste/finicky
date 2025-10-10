package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreServices
#include <stdlib.h>
#include "main.h"
*/
import "C"

import (
	"embed"
	"encoding/base64"
	"encoding/json"
	"finicky/browser"
	"finicky/config"
	"finicky/logger"
	"finicky/shorturl"
	"finicky/version"
	"finicky/window"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/dop251/goja"
)

//go:embed assets/finickyConfigAPI.js
var embeddedFiles embed.FS

type ProcessInfo struct {
	Name     string `json:"name"`
	BundleID string `json:"bundleId"`
	Path     string `json:"path"`
}

type UpdateInfo struct {
	ReleaseInfo        *version.ReleaseInfo
	UpdateCheckEnabled bool
}

type URLInfo struct {
	URL              string
	Opener           *ProcessInfo
	OpenInBackground bool
}

type ConfigInfo struct {
	Handlers       int16
	Rewrites       int16
	DefaultBrowser string
	ConfigPath     string
}

// FIXME: Clean up app global stae
var urlListener chan URLInfo = make(chan URLInfo)
var windowClosed chan struct{} = make(chan struct{})
var vm *config.VM

var forceWindowOpen bool = false
var queueWindowOpen chan bool = make(chan bool)
var lastError error
var dryRun bool = false
var updateInfo UpdateInfo
var configInfo *ConfigInfo
var currentConfigState *config.ConfigState
var shouldKeepRunning bool = true

func main() {
	startTime := time.Now()
	logger.Setup()
	runtime.LockOSThread()

	// Define command line flags
	configPathPtr := flag.String("config", "", "Path to custom configuration file")
	windowPtr := flag.Bool("window", false, "Force window to open")
	dryRunPtr := flag.Bool("dry-run", false, "Simulate without actually opening browsers")
	flag.Parse()

	// Use the parsed values
	customConfigPath := *configPathPtr
	if customConfigPath != "" {
		slog.Debug("Using custom config path", "path", customConfigPath)
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
	configChange := make(chan struct{})
	cfw, err := config.NewConfigFileWatcher(customConfigPath, namespace, configChange)

	if err != nil {
		handleFatalError(fmt.Sprintf("Failed to setup config file watcher: %v", err))
	}

	vm, err = setupVM(cfw, embeddedFiles, namespace)
	if err != nil {
		handleFatalError(err.Error())
	}

	slog.Debug("VM setup complete", "duration", fmt.Sprintf("%.2fms", float64(time.Since(startTime).Microseconds())/1000))

	go checkForUpdates()

	// Set up test URL handler
	window.TestUrlHandler = func(url string) {
		go TestURLInternal(url)
	}

	const oneDay = 24 * time.Hour

	var showingWindow bool = false
	timeoutChan := time.After(1 * time.Second)
	updateChan := time.After(oneDay)

	shouldKeepRunning = getConfigOption("keepRunning", true)
	if shouldKeepRunning {
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

				var browserConfig *browser.BrowserConfig
				var err error

				if vm != nil {
					browserConfig, err = evaluateURL(vm.Runtime(), url, urlInfo.Opener)
					if err != nil {
						handleRuntimeError(err)
					}
				} else {
					slog.Warn("No configuration available, using default configuration")
				}

				if browserConfig == nil {
					browserConfig = &browser.BrowserConfig{
						Name:             "com.apple.Safari",
						AppType:          "bundleId",
						OpenInBackground: &urlInfo.OpenInBackground,
						Profile:          "",
						Args:             []string{},
						URL:              url,
					}
				}

				if err := browser.LaunchBrowser(*browserConfig, dryRun, urlInfo.OpenInBackground); err != nil {
					slog.Error("Failed to start browser", "error", err)
				}

				slog.Debug("Time taken evaluating URL and opening browser", "duration", fmt.Sprintf("%.2fms", float64(time.Since(startTime).Microseconds())/1000))

				if !showingWindow && !shouldKeepRunning {
					timeoutChan = time.After(2 * time.Second)
				} else {
					timeoutChan = nil
				}

			case <-configChange:
				startTime := time.Now()
				var setupErr error
				slog.Debug("Config has changed")
				vm, setupErr = setupVM(cfw, embeddedFiles, namespace)
				if setupErr != nil {
					handleRuntimeError(setupErr)
				}
				slog.Debug("VM refresh complete", "duration", fmt.Sprintf("%.2fms", float64(time.Since(startTime).Microseconds())/1000))
				shouldKeepRunning = getConfigOption("keepRunning", true)

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
				if !shouldKeepRunning {
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

	hideIcon := getConfigOption("hideIcon", false)

	C.RunApp(C.bool(forceWindowOpen), C.bool(!hideIcon), C.bool(shouldKeepRunning))
}

func handleRuntimeError(err error) {
	slog.Error("Failed evaluating url", "error", err)
	lastError = err
	go QueueWindowDisplay(1)
}

func getConfigOption(optionName string, defaultValue bool) bool {
	if vm == nil || vm.Runtime() == nil {
		slog.Debug("VM not initialized, returning default for config option", "option", optionName, "default", defaultValue)
		return defaultValue
	}

	script := fmt.Sprintf("finickyConfigAPI.getOption('%s', finalConfig, %t)", optionName, defaultValue)
	optionVal, err := vm.Runtime().RunString(script)

	if err != nil {
		slog.Error("Failed to get config option", "option", optionName, "error", err)
		return defaultValue
	}

	return optionVal.ToBoolean()
}

//export HandleURL
func HandleURL(url *C.char, name *C.char, bundleId *C.char, path *C.char, openInBackground C.bool) {
	var opener ProcessInfo

	if name != nil && bundleId != nil && path != nil {
		opener = ProcessInfo{
			Name:     C.GoString(name),
			BundleID: C.GoString(bundleId),
			Path:     C.GoString(path),
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

//export TestURL
func TestURL(url *C.char) {
	urlString := C.GoString(url)
	TestURLInternal(urlString)
}

func TestURLInternal(urlString string) {
	slog.Debug("Testing URL", "url", urlString)

	if vm == nil {
		slog.Error("VM not initialized")
		window.SendMessageToWebView("testUrlResult", map[string]interface{}{
			"error": "Configuration not loaded",
		})
		return
	}

	browserConfig, err := evaluateURL(vm.Runtime(), urlString, nil)
	if err != nil {
		slog.Error("Failed to evaluate URL", "error", err)
		window.SendMessageToWebView("testUrlResult", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	if browserConfig == nil {
		window.SendMessageToWebView("testUrlResult", map[string]interface{}{
			"error": "No browser config returned",
		})
		return
	}

	window.SendMessageToWebView("testUrlResult", map[string]interface{}{
		"url":              browserConfig.URL,
		"browser":          browserConfig.Name,
		"openInBackground": browserConfig.OpenInBackground,
		"profile":          browserConfig.Profile,
		"args":             browserConfig.Args,
	})
}

func evaluateURL(vm *goja.Runtime, url string, opener *ProcessInfo) (*browser.BrowserConfig, error) {
	resolvedURL, err := shorturl.ResolveURL(url)
	vm.Set("originalUrl", url)

	if err != nil {
		// Continue with original URL if resolution fails
		slog.Info("Failed to resolve short URL", "error", err, "url", url, "using", resolvedURL)
	}

	url = resolvedURL

	vm.Set("url", resolvedURL)

	if opener != nil {
		vm.Set("opener", map[string]interface{}{
			"name":     opener.Name,
			"bundleId": opener.BundleID,
			"path":     opener.Path,
		})
		slog.Debug("Setting opener", "name", opener.Name, "bundleId", opener.BundleID, "path", opener.Path)
	} else {
		vm.Set("opener", nil)
		slog.Debug("No opener detected")
	}

	openResult, err := vm.RunString("finickyConfigAPI.openUrl(url, opener, originalUrl, finalConfig)")
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate URL in config: %v", err)
	}

	resultJSON := openResult.ToObject(vm).Export()
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

func handleFatalError(errorMessage string) {
	slog.Error("Fatal error", "msg", errorMessage)
	lastError = fmt.Errorf("%s", errorMessage)
	forceWindowOpen = true
}

//export QueueWindowDisplay
func QueueWindowDisplay(openWindow int32) {
	queueWindowOpen <- openWindow != 0
}

//export ShowConfigWindow
func ShowConfigWindow() {
	slog.Debug("Showing window")
	window.ShowWindow()

	// Send version information
	currentVersion := version.GetCurrentVersion()
	window.SendMessageToWebView("version", currentVersion)

}

//export WindowDidClose
func WindowDidClose() {
	windowClosed <- struct{}{}
}

//export GetCurrentConfigPath
func GetCurrentConfigPath() *C.char {
	if configInfo != nil && configInfo.ConfigPath != "" {
		cPath := C.CString(configInfo.ConfigPath)
		return cPath
	} else {
		return nil
	}
}

func checkForUpdates() {
	var runtime *goja.Runtime
	if vm != nil {
		runtime = vm.Runtime()
	}

	releaseInfo, updateCheckEnabled, err := version.CheckForUpdatesIfEnabled(runtime)
	if err != nil {
		slog.Error("Error checking for updates", "error", err)
	}

	updateInfo = UpdateInfo{
		ReleaseInfo:        releaseInfo,
		UpdateCheckEnabled: updateCheckEnabled,
	}

	if updateInfo.ReleaseInfo != nil && updateInfo.ReleaseInfo.HasUpdate {
		slog.Info("New version is available", "version", updateInfo.ReleaseInfo.LatestVersion)
	}

	if updateInfo.ReleaseInfo != nil {
		window.SendMessageToWebView("updateInfo", map[string]interface{}{
			"version":            updateInfo.ReleaseInfo.LatestVersion,
			"hasUpdate":          updateInfo.ReleaseInfo.HasUpdate,
			"updateCheckEnabled": updateInfo.UpdateCheckEnabled,
			"downloadUrl":        updateInfo.ReleaseInfo.DownloadUrl,
			"releaseUrl":         updateInfo.ReleaseInfo.ReleaseUrl,
		})
	} else {
		window.SendMessageToWebView("updateInfo", map[string]interface{}{
			"version":            "",
			"hasUpdate":          false,
			"updateCheckEnabled": updateInfo.UpdateCheckEnabled,
			"downloadUrl":        "",
			"releaseUrl":         "",
		})
	}
}

func tearDown() {
	checkForUpdates()
	slog.Info("Exiting...")
	os.Exit(0)
}

func setupVM(cfw *config.ConfigFileWatcher, embeddedFS embed.FS, namespace string) (*config.VM, error) {
	logRequests := true
	var err error

	defer func() {
		err = logger.SetupFile(logRequests)
		if err != nil {
			slog.Warn("Failed to setup file logging", "error", err)
		}
	}()

	currentBundlePath, configPath, err := cfw.BundleConfig()

	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	if currentBundlePath != "" {
		vm, err = config.New(embeddedFS, namespace, currentBundlePath)

		if err != nil {
			return nil, fmt.Errorf("failed to setup VM: %v", err)
		}

		currentConfigState = vm.GetConfigState()

		if currentConfigState != nil {
			configInfo = &ConfigInfo{
				Handlers:       currentConfigState.Handlers,
				Rewrites:       currentConfigState.Rewrites,
				DefaultBrowser: currentConfigState.DefaultBrowser,
				ConfigPath:     configPath,
			}
		}

		keepRunning := getConfigOption("keepRunning", true)
		hideIcon := getConfigOption("hideIcon", false)
		logRequests = getConfigOption("logRequests", false)
		checkForUpdates := getConfigOption("checkForUpdates", true)

		window.SendMessageToWebView("config", map[string]interface{}{
			"handlers":       configInfo.Handlers,
			"rewrites":       configInfo.Rewrites,
			"defaultBrowser": configInfo.DefaultBrowser,
			"configPath":     configInfo.ConfigPath,
			"options": map[string]interface{}{
				"keepRunning":     keepRunning,
				"hideIcon":        hideIcon,
				"logRequests":     logRequests,
				"checkForUpdates": checkForUpdates,
			},
		})

		return vm, nil
	}

	return nil, nil
}
