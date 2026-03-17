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
	Handlers       int16
	Rewrites       int16
	DefaultBrowser string
	ConfigPath     string
}

var urlListener chan URLInfo = make(chan URLInfo)
var windowClosed chan struct{} = make(chan struct{})
var vm *config.VM

var forceWindowOpen bool = false
var queueWindowOpen chan bool = make(chan bool)
var lastError error
var dryRun bool = false
var updateInfo UpdateInfo
var configInfo *ConfigInfo
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

	vm, err = setupVM(cfw, namespace)
	if err != nil {
		handleFatalError(err.Error())
	}

	slog.Debug("VM setup complete", "duration", fmt.Sprintf("%.2fms", float64(time.Since(startTime).Microseconds())/1000))

	go checkForUpdates()

	// Set up test URL handler
	window.TestUrlHandler = func(url string) {
		go TestURLInternal(url)
	}

	// Set up rules save handler.
	// When there is no JS config, rebuild the VM from the updated rules.
	// When there is a JS config, JSON rules are loaded fresh in evaluateURL — nothing to do.
	window.SaveRulesHandler = func(rf rules.RulesFile) {
		slog.Debug("Rules updated", "count", len(rf.Rules))
		resolver.SetCachedRules(rf)
		if vm == nil || !vm.IsJSConfig() {
			if rf.DefaultBrowser == "" && len(rf.Rules) == 0 {
				vm = nil
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
			vm = newVM
			shouldKeepRunning = vm.GetAllConfigOptions().KeepRunning
		}
	}

	const oneDay = 24 * time.Hour

	var showingWindow bool = false
	timeoutChan := time.After(1 * time.Second)
	updateChan := time.After(oneDay)

	shouldKeepRunning = vm.GetAllConfigOptions().KeepRunning
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

				config, err := resolver.ResolveURL(vm, url, urlInfo.Opener, urlInfo.OpenInBackground)
				if err != nil {
					handleRuntimeError(err)
				} else {
					lastError = nil
				}
				if launchErr := browser.LaunchBrowser(*config, dryRun, urlInfo.OpenInBackground); launchErr != nil {
					slog.Error("Failed to start browser", "error", launchErr)
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
				vm, setupErr = setupVM(cfw, namespace)
				if setupErr != nil {
					handleRuntimeError(setupErr)
				}
				slog.Debug("VM refresh complete", "duration", fmt.Sprintf("%.2fms", float64(time.Since(startTime).Microseconds())/1000))
				shouldKeepRunning = vm.GetAllConfigOptions().KeepRunning

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

	C.RunApp(C.bool(forceWindowOpen), C.bool(!vm.GetAllConfigOptions().HideIcon), C.bool(shouldKeepRunning))
}

func handleRuntimeError(err error) {
	slog.Error("Failed evaluating url", "error", err)
	lastError = err
	go QueueWindowDisplay(1)
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

//export TestURL
func TestURL(url *C.char) {
	urlString := C.GoString(url)
	TestURLInternal(urlString)
}

func TestURLInternal(urlString string) {
	slog.Debug("Testing URL", "url", urlString)

	config, err := resolver.ResolveURL(vm, urlString, nil, false)
	if err != nil {
		slog.Error("Failed to evaluate URL", "error", err)
		window.SendMessageToWebView("testUrlResult", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	window.SendMessageToWebView("testUrlResult", map[string]interface{}{
		"url":              config.URL,
		"browser":          config.Name,
		"openInBackground": config.OpenInBackground,
		"profile":          config.Profile,
		"args":             config.Args,
	})
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

func setupVM(cfw *config.ConfigFileWatcher, namespace string) (*config.VM, error) {
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
	if cs != nil {
		configInfo = &ConfigInfo{
			Handlers:       cs.Handlers,
			Rewrites:       cs.Rewrites,
			DefaultBrowser: cs.DefaultBrowser,
			ConfigPath:     configPath,
		}
	}

	opts := newVM.GetAllConfigOptions()
	logRequests = opts.LogRequests

	window.SendMessageToWebView("config", map[string]interface{}{
		"handlers":       configInfo.Handlers,
		"rewrites":       configInfo.Rewrites,
		"defaultBrowser": configInfo.DefaultBrowser,
		"configPath":     configInfo.ConfigPath,
		"isJSConfig":     newVM.IsJSConfig(),
		"options": map[string]interface{}{
			"keepRunning":     opts.KeepRunning,
			"hideIcon":        opts.HideIcon,
			"logRequests":     opts.LogRequests,
			"checkForUpdates": opts.CheckForUpdates,
		},
	})

	return newVM, nil
}
