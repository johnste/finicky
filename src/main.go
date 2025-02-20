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
	"log"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/dop251/goja"
)

//go:embed build/finickyConfigAPI.js
var embeddedFiles embed.FS

type ProcessInfo struct {
	Name     string
	BundleID string
	Path     string
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

func main() {
	startTime := time.Now()
	logger.Setup()
	if currentVersion := version.GetCurrentVersion(); currentVersion != "" {
		commitHash, buildDate := version.GetBuildInfo()
		log.Printf("Starting Finicky %s (built %s, commit %s)", currentVersion, buildDate, commitHash)
	} else {
		log.Println("Starting Finicky")
	}
	runtime.LockOSThread()


	go func() {
		is_default_browser, err := setDefaultBrowser()
		if err != nil {
			log.Printf("Failed checking if default browser is set: %v", err)
		} else if !is_default_browser {
			log.Println("We're not the default browser")
		} else {
			log.Println("We are the default browser")
		}
	}()

	vm, err := config.New(embeddedFiles)
	if err != nil {
		handleFatalError(fmt.Sprintf("Failed to setup VM: %v", err))
	}

	log.Printf("VM setup complete in %v", time.Since(startTime))

	go checkForUpdates()

	var showingWindow bool = false
	var timeoutChan = time.After(5 * time.Second)
	go func() {
		log.Println("Listening for events...")
		for {
			select {
			case urlInfo := <-urlListener:
				startTime := time.Now()

				log.Printf("URL received: %s", urlInfo.URL)

				var runtime *goja.Runtime
				if vm != nil {
					runtime = vm.Runtime()
				}

				err := evaluateURL(runtime, urlInfo.URL, urlInfo.Opener)
				if err != nil {
					log.Printf("Failed to load config: %v", err)

				}

				elapsedTime := time.Since(startTime)
				log.Printf("Time taken evaluating URL and opening browser: %v", elapsedTime)

				if !showingWindow {
					timeoutChan = time.After(5 * time.Second)
				} else {
					timeoutChan = nil
				}

			case shouldShowWindow := <-queueWindowOpen:
				if !showingWindow && shouldShowWindow {
					go ShowTheMainWindow(err)
					showingWindow = true
					timeoutChan = nil
				}

			case <-windowClosed:
				log.Println("Exiting due to window closed")
				tearDown()

			case <-timeoutChan:
				log.Println("Exiting due to timeout")
				tearDown()
			}
		}
	}()

	C.RunApp(C.int(forceWindowOpen))
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

func evaluateURL(vm *goja.Runtime, url string, opener *ProcessInfo) error {
	resolvedURL, err := shorturl.ResolveURL(url)
	if err != nil {
		// Continue with original URL if resolution fails
		log.Printf("Failed to resolve short URL: %v", err)

	} else {
		url = resolvedURL
	}

	if vm == nil {
		log.Printf("VM is not properly initialized, skipping evaluation and defaulting to Safari")
		browserConfig := browser.BrowserConfig{
			Name: "com.apple.Safari",
			AppType: "bundleId",
			OpenInBackground: false,
			Profile: "",
			Args: []string{},
			URL: url,
		}
		return browser.LaunchBrowser(browserConfig)
	}

	vm.Set("url", resolvedURL)
	vm.Set("opener", opener)

	openResult, err := vm.RunString("finickyConfigAPI.openUrl(url, opener, finalConfig)")
	if err != nil {
		handleFatalError(fmt.Sprintf("Failed to evaluate URL in config: %v", err))
		return err
	}

	resultJSON := openResult.ToObject(vm).Export()
	resultBytes, err := json.Marshal(resultJSON)
	if err != nil {
		handleFatalError(fmt.Sprintf("Failed to process browser configuration: %v", err))
		return err
	}

	var browserConfig browser.BrowserConfig

	if err := json.Unmarshal(resultBytes, &browserConfig); err != nil {
		handleFatalError(fmt.Sprintf("Failed to parse browser configuration: %v", err))
		return err
	}

	log.Printf("Final browser options: %+v", browserConfig)

	err = browser.LaunchBrowser(browserConfig)
	if err != nil {
		log.Printf("Failed to start browser: %v", err)
		return err
	}

	return nil
}

func handleFatalError(errorMessage string) {

	log.Printf("Fatal error: %s", errorMessage)
	forceWindowOpen = 1
}

//export QueueWindowDisplay
func QueueWindowDisplay(openWindow int32) {
	queueWindowOpen <- openWindow != 0
}

func ShowTheMainWindow(err error) {
	log.Println("Showing window")
	window.ShowWindow()

	if err != nil {
		window.SendMessageToWebView("status", "🔴 An error occurred")
	} else {
		window.SendMessageToWebView("status", "🟢 Everything should be good")
	}

	if vm != nil {
		configInfo, err := vm.Runtime().RunString("finickyConfigAPI.getConfigInfo(finalConfig)")
		if err != nil {
			log.Printf("failed to get config info: %v", err)
		} else {
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
	log.Println("Window closed, exiting")
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
		log.Printf("Error checking for updates: %v", err)
	}
}

func tearDown() {
	log.Println("Exiting...")
	os.Exit(0)
}