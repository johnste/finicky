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
	"finicky/util"
	"finicky/version"
	"log"
	"os"
	"runtime"
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
	PID    int32
	Opener *ProcessInfo
}

var urlListener chan URLInfo = make(chan URLInfo)

func main() {
	startTime := time.Now()
	logger.Setup()
	log.Println("Starting Finicky 🍇")
	runtime.LockOSThread()

	is_default_browser, err := setDefaultBrowser()
	if err != nil {
		log.Printf("failed checking if default browser is set: %v", err)
	}

	if !is_default_browser {
		log.Println("Is not the default browser")
	} else {
		log.Println("Is the default browser")
	}

	vm, err := config.New(embeddedFiles)
	if err != nil {
		log.Fatalf("failed to setup VM: %v", err)
	}

	// Set system-specific functions
	vm.SetModifierKeysFunc(util.GetModifierKeys)
	vm.SetSystemInfoFunc(util.GetSystemInfo)

	log.Printf("VM setup complete in %v", time.Since(startTime))

	go func() {
		for {
			log.Println("Listening for URL...")

			select {
			case urlInfo := <-urlListener:
				startTime := time.Now()
				url := urlInfo.URL
				opener := urlInfo.Opener
				pid := urlInfo.PID

				log.Printf("URL received! %s", url)
				err = evaluateURL(vm.Runtime(), url, pid, opener)
				if err != nil {
					log.Printf("failed to load config: %v", err)
					os.Exit(1)
				}

				elapsedTime := time.Since(startTime)
				log.Printf("Time taken: %v", elapsedTime)

			case <-time.After(15 * time.Second):
				if err := version.CheckForUpdatesFromConfig(vm.Runtime()); err != nil {
					log.Printf("Error checking for updates: %v", err)
				}
				os.Exit(1)

			case <-time.After(15 * time.Second):
				log.Println("Exiting all processes due to timeout")
				os.Exit(1)
			}
		}
	}()

	C.RunApp()
}

//export HandleURL
func HandleURL(url *C.char, name *C.char, bundleId *C.char, path *C.char, pid C.int) {
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
		PID:    int32(pid),
		Opener: &opener,
	}
}

func evaluateURL(vm *goja.Runtime, url string, pid int32, opener *ProcessInfo) error {
	resolvedURL, err := shorturl.ResolveURL(url)
	if err != nil {
		log.Printf("Failed to resolve short URL: %v", err)
		// Continue with original URL if resolution fails
	} else {
		url = resolvedURL
	}

	vm.Set("url", resolvedURL)
	vm.Set("opener", opener)
	vm.Set("pid", pid)

	openResult, err := vm.RunString("finickyConfigAPI.openUrl(url, pid, opener, mergedConfig)")
	if err != nil {
		log.Fatalf("failed to get result: %v", err)
	}

	var browserConfig browser.BrowserConfig
	resultJSON := openResult.ToObject(vm).Export()
	resultBytes, err := json.Marshal(resultJSON)
	if err != nil {
		log.Fatalf("failed to marshal result: %v", err)
	}

	if err := json.Unmarshal(resultBytes, &browserConfig); err != nil {
		log.Fatalf("failed to unmarshal browser config: %v", err)
	}

	log.Printf("Final browser options: %+v", browserConfig)

	err = browser.LaunchBrowser(browserConfig)
	if err != nil {
		log.Printf("Failed to start browser: %v", err)
		return err
	}

	return nil
}