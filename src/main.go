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
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/dop251/goja"
)

//go:embed build/finickyConfigAPI.js
var embeddedFiles embed.FS

type ProcessInfo struct {
	Name     string `json:"name"`
	BundleID string `json:"bundleId"`
	Path     string `json:"path"`
}

type URLInfo struct {
	URL    string
	Opener *ProcessInfo
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

func main() {

	startTime := time.Now()
	logger.Setup()
	runtime.LockOSThread()

	customConfigPath := ""
	for _, arg := range os.Args {
		slog.Debug("Processing argument", "arg", arg)
		if strings.HasPrefix(arg, "--config=") {
			customConfigPath = strings.TrimPrefix(arg, "--config=")
			slog.Debug("Using custom config path", "path", customConfigPath)
		}
		if strings.HasPrefix(arg, "--window") {
			forceWindowOpen = 1
		}
		if strings.HasPrefix(arg, "--dry-run") {
			dryRun = true
		}
	}

	if currentVersion := version.GetCurrentVersion(); currentVersion != "" {
		commitHash, buildDate := version.GetBuildInfo()
		slog.Info("Starting Finicky", "version", currentVersion, "buildDate", buildDate, "commitHash", commitHash)
	} else {
		slog.Info("Starting Finicky")
	}

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

	var err error
	vm, err = config.New(embeddedFiles, customConfigPath)
	if err != nil {
		handleFatalError(fmt.Sprintf("Failed to setup VM: %v", err))
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

				var runtime *goja.Runtime
				if vm != nil {
					runtime = vm.Runtime()
				}

				browserConfig, err := evaluateURL(runtime, urlInfo.URL, urlInfo.Opener)
				if err != nil {
					handleRuntimeError(err)
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

	if vm == nil {
		return nil, fmt.Errorf("Can't access config")
	}

	vm.Set("url", resolvedURL)
	vm.Set("opener", opener)

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

	slog.Debug("Final browser options", "options", fmt.Sprintf("%+v", browserResult.Browser))
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

	if err != nil {
		window.SendMessageToWebView("status", "🔴 An error occurred")
	} else {
		window.SendMessageToWebView("status", "🟢 Everything should be good")
	}

	if vm != nil {
		configInfo, err := vm.Runtime().RunString("finickyConfigAPI.getConfigInfo(finalConfig)")
		if err == nil {
			window.SendMessageToWebView("configInfo", configInfo.String())
		}
	} else {
		window.SendMessageToWebView("configInfo", "Configuration was not properly loaded")
	}

	// Send version information
	currentVersion := version.GetCurrentVersion()
	if currentVersion != "" {
		window.SendMessageToWebView("version", currentVersion)
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

	if err := version.CheckForUpdatesFromConfig(runtime); err != nil {
		slog.Error("Error checking for updates", "error", err)
	}
}

func tearDown() {
	checkForUpdates()
	slog.Info("Exiting...")
	os.Exit(0)
}