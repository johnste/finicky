package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework CoreServices
#include <stdlib.h>
#include "main.h"
*/
import "C"

import (
	"bytes"
	"embed"
	"encoding/json"
	"finicky/browser"
	"finicky/shorturl"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"time"

	"github.com/dop251/goja"
	"github.com/evanw/esbuild/pkg/api"
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
var file *os.File

var namespace = "finickyConfig"

func main() {
	startTime := time.Now()
	setupLogger()
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

	vm, err := setupVM()
	if err != nil {
		log.Fatalf("failed to setup VM: %v", err)
	}

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
				err = evaluateURL(vm, url, pid, opener)
				if err != nil {
					log.Printf("failed to load config: %v", err)
					os.Exit(1)
				}

				elapsedTime := time.Since(startTime)
				log.Printf("Time taken: %v", elapsedTime)

			case <-time.After(5 * time.Second):
				log.Println("Exiting all processes due to timeout")
				os.Exit(1)
			}
		}
	}()

	C.RunApp()

}

func setupVM() (*goja.Runtime, error) {
	bundlePath, simpleConfigPath, err := prepareConfig()
	if err != nil {
		log.Fatalf("failed to bundle config: %v", err)
	}
	apiContent, err := embeddedFiles.ReadFile("build/finickyConfigAPI.js")
	if err != nil {
		log.Fatalf("failed to read bundled file: %v", err)
	}
	var content []byte
	if bundlePath != "" {
		content, err = os.ReadFile(bundlePath)
		if err != nil {
			log.Fatalf("failed to read file: %v", err)
		}
	}
	var simpleContent []byte
	if simpleConfigPath != "" {
		simpleContent, err = os.ReadFile(simpleConfigPath)
		var simpleConfig map[string]interface{}
		if err := json.Unmarshal(simpleContent, &simpleConfig); err != nil {
			log.Fatalf("failed to unmarshal simple config: %v", err)
		}
		if err != nil {
			log.Fatalf("failed to read file: %v", err)
		}
	}
	vm := goja.New()
	vm.Set("self", vm.GlobalObject())
	vm.Set("console", GetConsoleMap())
	log.Println("Evaluating API script...")
	_, err = vm.RunString(string(apiContent))
	if err != nil {
		log.Fatalf("failed to run api script: %v", err)
	}
	log.Println("Done evaluating API script")
	userAPI := vm.Get("finickyConfigAPI").ToObject(vm).Get("utilities").ToObject(vm)
	finicky := make(map[string]interface{})
	for _, key := range userAPI.Keys() {
		finicky[key] = userAPI.Get(key)
	}

	finicky["getModifierKeys"] = getModifierKeys
	finicky["getSystemInfo"] = getSystemInfo
	vm.Set("finicky", finicky)
	if content != nil {
		_, err = vm.RunString(string(content))
		if err != nil {
			return nil, fmt.Errorf("failed to run config script: %v", err)
		}
	} else {
		vm.Set(namespace, map[string]interface{}{})
	}
	mergedConfig, err := vm.RunString(fmt.Sprintf("finickyConfigAPI.mergeConfig(%s.default, %s)", namespace, simpleContent))
	if err != nil {
		return nil, fmt.Errorf("failed to get merged config: %v", err)
	}
	vm.Set("mergedConfig", mergedConfig)
	validConfig, err := vm.RunString("finickyConfigAPI.validateConfig(mergedConfig)")
	if err != nil {
		return nil, fmt.Errorf("failed to get valid config: %v", err)
	}
	if !validConfig.ToBoolean() {
		return nil, fmt.Errorf("configuration is invalid: %s", validConfig.String())
	}

	// Check logRequests option
	logRequests, err := vm.RunString("finickyConfigAPI.getOption('logRequests', mergedConfig)")
	log.Printf("logRequests: %v", logRequests)

	// Setup file logging if enabled
	if err := setupFileLogger(logRequests.ToBoolean()); err != nil {
		log.Printf("Warning: Failed to setup file logging: %v", err)
	}

	return vm, nil
}

//export HandleURL
func HandleURL(url *C.char, name *C.char, bundleID *C.char, path *C.char, pid C.int) {
	var opener ProcessInfo

	if name != nil && bundleID != nil && path != nil {
		opener = ProcessInfo{
			Name:     C.GoString(name),
			BundleID: C.GoString(bundleID),
			Path:     C.GoString(path),
		}
	}

	urlListener <- URLInfo{
		URL:    C.GoString(url),
		PID:    int32(pid),
		Opener: &opener,
	}
}

func prepareConfig() (string, string, error) {
	simpleConfigPath, errSimple := getSimpleConfigPath()
	configPath, err := getConfigPath()
	if err != nil && errSimple != nil {
		return "", "", fmt.Errorf("failed to get config or simple config: %v", err)
	}

	if configPath == "" {
		log.Printf("Found config path: <None>")
	} else {
		log.Printf("Found config path: %s", configPath)
	}
	if simpleConfigPath == "" {
		log.Printf("Found simple config path: <None>")
	} else {
		log.Printf("Found simple config path: %s", simpleConfigPath)
	}

	if configPath != "" {
		bundlePath := os.TempDir() + "/finicky_output.js"

		result := api.Build(api.BuildOptions{
			EntryPoints: []string{configPath},
			Outfile:     bundlePath,
			Bundle:      true,
			Write:       true,
			LogLevel:    api.LogLevelError,
			Platform:    api.PlatformNeutral,
			Target:      api.ES2015,
			Format:      api.FormatIIFE,
			GlobalName:  namespace,
		})

		if len(result.Errors) > 0 {
			log.Printf("Build errors: %v", result.Errors)
			os.Exit(1)
		}
		return bundlePath, simpleConfigPath, nil
	}

	return "", simpleConfigPath, nil
}

func getConfigPath() (string, error) {

	var configPaths []string
	if os.Getenv("DEBUG") == "true" {
		configPaths = append(configPaths, "./test/example.js")
	} else {
		configPaths = append(configPaths,
			"$HOME/.finicky.js",
			"$HOME/.config/.finicky.js",
		)
	}

	for _, path := range configPaths {
		expandedPath := os.ExpandEnv(path)
		if _, err := os.Stat(expandedPath); err == nil {
			return expandedPath, nil
		}
	}

	return "", fmt.Errorf("no config file found")
}

func getSimpleConfigPath() (string, error) {
	var configPaths []string
	if os.Getenv("DEBUG") == "true" {
		configPaths = append(configPaths, "./test/example.json")
	} else {
		configPaths = append(configPaths,
			"$HOME/.finicky.json",
			"$HOME/.config/.finicky.json",
		)
	}

	for _, path := range configPaths {
		expandedPath := os.ExpandEnv(path)
		if _, err := os.Stat(expandedPath); err == nil {
			return expandedPath, nil
		}
	}

	return "", fmt.Errorf("no simple config file found")
}

func evaluateURL(vm *goja.Runtime, url string, pid int32, opener *ProcessInfo) error {

	log.Printf("Evaluating URL: %s, PID: %d, Opener: %+v", url, pid, opener)

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

func setupLogger() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// Start with in-memory logging
	memLog = &bytes.Buffer{}
	multiWriter = io.MultiWriter(memLog, os.Stdout)
	log.SetOutput(multiWriter)
}

func setupFileLogger(shouldLog bool) error {
	if !shouldLog {
		return nil
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %v", err)
	}

	logDir := filepath.Join(homeDir, "Library", "Logs", "Finicky")
	err = os.MkdirAll(logDir, 0755) // Create directory if it doesn't exist
	if err != nil {
		return fmt.Errorf("failed to create log directory: %v", err)
	}

	currentTime := time.Now().Format("2006-01-02_15-04-05.000")
	logFile := filepath.Join(logDir, fmt.Sprintf("Finicky_%s.log", currentTime))

	file, err = os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// Write buffered logs to file
	if _, err := file.Write(memLog.Bytes()); err != nil {
		return fmt.Errorf("failed to write buffered logs: %v", err)
	}

	// Update writer to include file
	multiWriter = io.MultiWriter(file, os.Stdout)
	log.SetOutput(multiWriter)
	return nil
}

func closeLogger() {
	log.Println("Application closed!")
	if file != nil {
		defer file.Close()
	}
}

func getModifierKeys() map[string]bool {
	keys := C.getModifierKeys()
	return map[string]bool{
		"shift":    bool(keys.shift),
		"option":   bool(keys.option),
		"command":  bool(keys.command),
		"control":  bool(keys.control),
		"capsLock": bool(keys.capsLock),
		"fn":       bool(keys.fn),
	}
}

func getSystemInfo() map[string]string {
	info := C.getSystemInfo()
	return map[string]string{
		"localizedName": C.GoString(info.localizedName),
		"name":          C.GoString(info.name),
	}
}
