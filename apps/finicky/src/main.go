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
	ReleaseInfo *version.ReleaseInfo
	UpdateCheckEnabled bool
}

type URLInfo struct {
	URL    string
	Opener *ProcessInfo
}

type ConfigInfo struct {
	Handlers int16
	Rewrites int16
	DefaultBrowser string
	ConfigPath string
}

// FIXME: Clean up app global stae
var urlListener chan URLInfo = make(chan URLInfo)
var windowClosed chan struct{} = make(chan struct{})
var vm *config.VM
// FIXME: find a better data type for this
var forceWindowOpen int32 = 0
var queueWindowOpen chan bool = make(chan bool)
var lastError error
var dryRun bool = false
var updateInfo UpdateInfo
var configInfo *ConfigInfo
var currentConfigState *config.ConfigState

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
		forceWindowOpen = 1
	}

	dryRun = *dryRunPtr

	currentVersion := version.GetCurrentVersion();
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

	var showingWindow bool = false
	var timeoutChan = time.After(1 * time.Second)
	go func() {
		slog.Info("Listening for events...")
		for {
			select {
			case urlInfo := <-urlListener:
				startTime := time.Now()

				slog.Info("URL received", "url", urlInfo.URL)

				var browserConfig *browser.BrowserConfig
				var err error

				if vm != nil {
					browserConfig, err = evaluateURL(vm.Runtime(), urlInfo.URL, urlInfo.Opener)
					if err != nil {
						handleRuntimeError(err)
					}
				} else {
					slog.Warn("No configuration available, using default configuration")
				}

				if browserConfig == nil {
					browserConfig = &browser.BrowserConfig{
						Name:            "com.apple.Safari",
						AppType:         "bundleId",
						OpenInBackground: false,
						Profile:         "",
						Args:            []string{},
						URL:             urlInfo.URL,
					}
				}

				if err := browser.LaunchBrowser(*browserConfig, dryRun); err != nil {
					slog.Error("Failed to start browser", "error", err)
				}

				slog.Debug("Time taken evaluating URL and opening browser", "duration", fmt.Sprintf("%.2fms", float64(time.Since(startTime).Microseconds())/1000))

				if !showingWindow {
					timeoutChan = time.After(2 * time.Second)
				} else {
					timeoutChan = nil
				}

			case <-configChange:
				startTime := time.Now()
				var setupErr error
				vm, setupErr = setupVM(cfw, embeddedFiles, namespace)
				if setupErr != nil {
					handleRuntimeError(setupErr)
				}
				slog.Debug("VM refresh complete", "duration", fmt.Sprintf("%.2fms", float64(time.Since(startTime).Microseconds())/1000))
				if configInfo != nil {
					if setupErr == nil {
						window.SendMessageToWebView("config", map[string]interface{}{
							"handlers": configInfo.Handlers,
							"rewrites": configInfo.Rewrites,
							"defaultBrowser": configInfo.DefaultBrowser,
							"configPath": configInfo.ConfigPath,
						})
					} else {
						window.SendMessageToWebView("config", map[string]interface{}{
							"handlers": 0,
							"rewrites": 0,
							"defaultBrowser": "",
							"configPath": configInfo.ConfigPath,
						})
					}

				}

			case shouldShowWindow := <-queueWindowOpen:
				if !showingWindow && shouldShowWindow {
					go ShowTheMainWindow(lastError)
					showingWindow = true
					timeoutChan = nil
				}

			case <-windowClosed:
				slog.Info("Exiting due to window closed")
				tearDown()

			case <-timeoutChan:
				slog.Info("Exiting due to timeout")
				tearDown()
			}
		}
	}()

	C.RunApp(C.int(forceWindowOpen))
}

func handleRuntimeError(err error) {
	slog.Error("Failed evaluating url", "error", err)
	lastError = err
	go QueueWindowDisplay(1)
}

//export HandleURL
func HandleURL(url *C.char, name *C.char, bundleId *C.char, path *C.char) {
	var opener ProcessInfo

	if name != nil && bundleId != nil && path != nil {
		opener = ProcessInfo{
			Name:     C.GoString(name),
			BundleID: C.GoString(bundleId),
			Path:     C.GoString(path),
		}
	}

	urlListener <- URLInfo{
		URL:    C.GoString(url),
		Opener: &opener,
	}
}

func evaluateURL(vm *goja.Runtime, url string, opener *ProcessInfo) (*browser.BrowserConfig, error) {
	resolvedURL, err := shorturl.ResolveURL(url)
	if err != nil {
		// Continue with original URL if resolution fails
		slog.Info("Failed to resolve short URL", "error", err)

	} else {
		url = resolvedURL
	}


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

	openResult, err := vm.RunString("finickyConfigAPI.openUrl(url, opener, finalConfig)")
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
	forceWindowOpen = 1
}

//export QueueWindowDisplay
func QueueWindowDisplay(openWindow int32) {
	queueWindowOpen <- openWindow != 0
}

func ShowTheMainWindow(err error) {
	slog.Debug("Showing window")
	window.ShowWindow()

	// Send version information
	currentVersion := version.GetCurrentVersion()
	window.SendMessageToWebView("version", currentVersion)


	slog.Debug("Update info", "updateInfo", updateInfo, "releaseInfo", updateInfo.ReleaseInfo)
	if updateInfo.ReleaseInfo != nil {
		window.SendMessageToWebView("updateInfo", map[string]interface{}{
			"version": updateInfo.ReleaseInfo.LatestVersion,
			"hasUpdate": updateInfo.ReleaseInfo.HasUpdate,
			"updateCheckEnabled": updateInfo.UpdateCheckEnabled,
		})
	} else {
		window.SendMessageToWebView("updateInfo", map[string]interface{}{
			"version": "",
			"hasUpdate": false,
			"updateCheckEnabled": updateInfo.UpdateCheckEnabled,
		})
	}

	if configInfo != nil {
		window.SendMessageToWebView("config", map[string]interface{}{
			"handlers": configInfo.Handlers,
			"rewrites": configInfo.Rewrites,
			"defaultBrowser": configInfo.DefaultBrowser,
			"configPath": configInfo.ConfigPath,
		})
	}

	// Send all buffered logs
	bufferedLogs := logger.GetBufferedLogs()
	for _, line := range strings.Split(bufferedLogs, "\n") {
		if line != "" {
			window.SendMessageToWebView("log", line)
		}
	}

	<-windowClosed
	slog.Info("Window closed, exiting")
	tearDown()
}

//export WindowDidClose
func WindowDidClose() {
	windowClosed <- struct{}{}
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
		ReleaseInfo: releaseInfo,
		UpdateCheckEnabled: updateCheckEnabled,
	}

	if updateInfo.ReleaseInfo != nil && updateInfo.ReleaseInfo.HasUpdate {
		slog.Info("New version is available", "version", updateInfo.ReleaseInfo.LatestVersion)
	}
}

func tearDown() {
	checkForUpdates()
	slog.Info("Exiting...")
	os.Exit(0)
}

func setupVM(cfw *config.ConfigFileWatcher, embeddedFS embed.FS, namespace string) (*config.VM, error) {
	shouldLogToFile := true
	var err error

	defer func() {
		err = logger.SetupFile(shouldLogToFile)
		if err != nil {
			slog.Warn("Failed to setup file logging", "error", err)
		}
	}()

	currentBundlePath, configPath, err := cfw.BundleConfig()

	if err != nil {
		return nil, fmt.Errorf("failed to read config: %v", err)
	}

	if currentBundlePath != "" {
		vm, err := config.New(embeddedFS, namespace, currentBundlePath)
		if err != nil {
			return nil, fmt.Errorf("failed to setup VM: %v", err)
		}

		// Update logging preference based on VM if available
		shouldLogToFile = vm.ShouldLogToFile(false)

		currentConfigState = vm.GetConfigState()

		if currentConfigState != nil {
			configInfo = &ConfigInfo{
				Handlers:       currentConfigState.Handlers,
				Rewrites:      currentConfigState.Rewrites,
				DefaultBrowser: currentConfigState.DefaultBrowser,
				ConfigPath:     strings.Replace(configPath, os.Getenv("HOME"), "~", 1),
			}
		}

		return vm, nil
	}

	return nil, nil
}



