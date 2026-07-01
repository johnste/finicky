//go:build windows

package main

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
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

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

// windowClosed is signalled by the UI loop once RunWindow returns. Buffered so
// the UI loop never blocks if the event loop is momentarily busy.
var windowClosed chan struct{} = make(chan struct{}, 1)

// showWindowChan carries show-window requests from the event loop (any
// goroutine) to the UI loop (the main, thread-locked goroutine that owns the
// WebView2 window). Buffered so requesters never block.
var showWindowChan chan struct{} = make(chan struct{}, 1)

// queueWindowOpen lets other goroutines (e.g. the IPC listener handling a
// second `--window` invocation) ask the event loop to open the window, keeping
// the event loop's showingWindow bookkeeping authoritative.
var queueWindowOpen chan bool = make(chan bool, 1)

var configChange chan struct{} = make(chan struct{}, 1)

var vm *config.VM

var forceWindowOpen bool = false
var lastError error
var dryRun bool = false
var skipJSConfig bool = false
var updateInfo UpdateInfo
var configInfo *ConfigInfo
var shouldKeepRunning bool = true

var (
	kernel32Mutex    = syscall.NewLazyDLL("kernel32.dll")
	procCreateMutexW = kernel32Mutex.NewProc("CreateMutexW")
)

const mutexName = "Global\\FinickyBrowserRouter"
const errorAlreadyExists = 183

// showWindowCmd is the sentinel a secondary instance sends over the IPC socket
// to ask the primary to open its config window (used when Finicky is launched a
// second time with --window, e.g. from the Start Menu entry). A real URL can
// never collide with this value.
const showWindowCmd = "__finicky_show_window__"

func main() {
	startTime := time.Now()
	logger.Setup()

	// Pin the main goroutine to the main OS thread. The WebView2 window is
	// created and pumped on this thread (see runUILoop / window.RunWindow);
	// Win32 requires the window's message loop to run on its creating thread.
	runtime.LockOSThread()

	flag.String("config", "", "Path to custom JS configuration file")
	rulesPathPtr := flag.String("rules", "", "Path to custom rules JSON file")
	noConfigPtr := flag.Bool("no-config", false, "Skip JS configuration file entirely")
	windowPtr := flag.Bool("window", false, "Force window to open")
	dryRunPtr := flag.Bool("dry-run", false, "Simulate without actually opening browsers")
	flag.Parse()

	if *windowPtr {
		forceWindowOpen = true
	}

	// URL passed as an argument (Windows protocol-handler invocation).
	urlFromArgs := ""
	if args := flag.Args(); len(args) > 0 {
		urlFromArgs = args[0]
	}

	// Single-instance detection via a named mutex. CreateMutexW succeeds even
	// when the mutex already exists, signalling that case through the error
	// value captured by Call (which reads GetLastError immediately after the
	// syscall). Reading GetLastError in a *separate* LazyProc.Call would race
	// with the Go runtime's own intervening syscalls, so we must use this one.
	mutexNamePtr, _ := syscall.UTF16PtrFromString(mutexName)
	handle, _, callErr := procCreateMutexW.Call(0, 0, uintptr(unsafe.Pointer(mutexNamePtr)))
	alreadyRunning := callErr == syscall.Errno(errorAlreadyExists)

	if alreadyRunning {
		// Another instance owns the routing. Hand our work to it and exit.
		if handle != 0 {
			syscall.CloseHandle(syscall.Handle(handle))
		}
		handOffToPrimary(urlFromArgs)
		os.Exit(0)
	}
	if handle != 0 {
		defer syscall.CloseHandle(syscall.Handle(handle))
	}

	customConfigPath := ""
	flag.Visit(func(f *flag.Flag) {
		if f.Name == "config" {
			customConfigPath = f.Value.String()
		}
	})

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

	dryRun = *dryRunPtr

	currentVersion := version.GetCurrentVersion()
	commitHash, buildDate := version.GetBuildInfo()
	slog.Info("Starting Finicky", "version", currentVersion)
	slog.Debug("Build info", "buildDate", buildDate, "commitHash", commitHash)

	go logDefaultBrowserStatus()

	namespace := "finickyConfig"
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

	window.TestUrlHandler = func(url string) {
		go TestURLInternal(url)
	}

	window.SaveRulesHandler = func(rf rules.RulesFile) {
		slog.Debug("Rules updated", "count", len(rf.Rules))
		resolver.SetCachedRules(rf)
		if vm == nil || !vm.IsJSConfig() {
			if rf.DefaultBrowser == "" && len(rf.Rules) == 0 && rf.Options == nil {
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
			if vm != nil {
				shouldKeepRunning = vm.GetAllConfigOptions().KeepRunning
				go checkForUpdates()
			}
		}
	}

	// Start the IPC server that receives URLs (and show-window requests) from
	// secondary instances.
	go listenForURLs()

	// If launched with a URL argument, process it.
	if urlFromArgs != "" {
		go func() {
			handleURLString(urlFromArgs, false)
		}()
	}

	if vm != nil {
		shouldKeepRunning = vm.GetAllConfigOptions().KeepRunning
	}

	// The event loop runs on a background goroutine (like the macOS build); the
	// main thread stays free to own the WebView2 UI in runUILoop.
	go runEventLoop(cfw, namespace)

	runUILoop()
}

// runUILoop owns the WebView2 window and MUST run on the main thread (pinned by
// runtime.LockOSThread in main). It waits for a show request, then creates and
// pumps the window's message loop here — blocking until the window closes —
// before notifying the event loop and looping back to wait for the next one.
func runUILoop() {
	for range showWindowChan {
		window.RunWindow()
		select {
		case windowClosed <- struct{}{}:
		default:
		}
	}
}

// requestShowWindow asks the UI loop (main thread) to open the window and seeds
// the current state, which buffers in the window package until the page loads.
func requestShowWindow() {
	window.SendMessageToWebView("version", version.GetCurrentVersion())
	select {
	case showWindowChan <- struct{}{}:
	default:
	}
}

func runEventLoop(cfw *config.ConfigFileWatcher, namespace string) {
	const oneDay = 24 * time.Hour

	showingWindow := false
	timeoutChan := time.After(1 * time.Second)
	updateChan := time.After(oneDay)

	if shouldKeepRunning {
		timeoutChan = nil
	}

	slog.Info("Listening for events...")

	if forceWindowOpen {
		requestShowWindow()
		showingWindow = true
		timeoutChan = nil
	}

	for {
		select {
		case urlInfo := <-urlListener:
			startTime := time.Now()
			slog.Info("URL received", "url", urlInfo.URL)

			config, err := resolver.ResolveURL(vm, urlInfo.URL, urlInfo.Opener, urlInfo.OpenInBackground)
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
			slog.Debug("Config has changed")
			var setupErr error
			vm, setupErr = setupVM(cfw, namespace)
			if setupErr != nil {
				handleRuntimeError(setupErr)
			} else {
				lastError = nil
			}
			slog.Debug("VM refresh complete", "duration", fmt.Sprintf("%.2fms", float64(time.Since(startTime).Microseconds())/1000))
			if vm != nil {
				shouldKeepRunning = vm.GetAllConfigOptions().KeepRunning
				go checkForUpdates()
			}

		case shouldShowWindow := <-queueWindowOpen:
			if !showingWindow && shouldShowWindow {
				requestShowWindow()
				showingWindow = true
				timeoutChan = nil
			}

		case <-updateChan:
			go checkForUpdates()
			updateChan = time.After(oneDay)

		case <-windowClosed:
			showingWindow = false
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
}

func handleURLString(urlString string, openInBackground bool) {
	if strings.HasPrefix(urlString, "finicky://open/") {
		encodedURL := strings.TrimPrefix(urlString, "finicky://open/")
		if decodedBytes, err := base64.StdEncoding.DecodeString(encodedURL); err == nil {
			urlString = string(decodedBytes)
			slog.Debug("Decoded finicky protocol URL", "decoded", urlString)
		} else {
			slog.Warn("Failed to decode finicky protocol URL", "error", err)
		}
	}

	urlListener <- URLInfo{
		URL:              urlString,
		Opener:           &resolver.OpenerInfo{},
		OpenInBackground: openInBackground,
	}
}

func handleRuntimeError(err error) {
	slog.Error("Failed evaluating url", "error", err)
	lastError = err
}

// logDefaultBrowserStatus reports whether Finicky is the default browser. On
// Windows 10+ an app cannot make itself default programmatically — the user
// does it in Settings (offered by the installer) — so unlike the macOS build we
// only observe and log here, never prompt. Prompting on every launch would pop
// the Settings app on every link click.
func logDefaultBrowserStatus() {
	isDefault, err := isDefaultBrowser()
	if err != nil {
		slog.Debug("Failed checking if we are the default browser", "error", err)
	} else if isDefault {
		slog.Debug("Finicky is the default browser")
	} else {
		slog.Debug("Finicky is not the default browser")
	}
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

	var currentBundlePath, configPath string
	if !skipJSConfig {
		var err2 error
		currentBundlePath, configPath, err2 = cfw.BundleConfig()
		if err2 != nil {
			return nil, fmt.Errorf("failed to read config: %v", err2)
		}
	}

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

// listenForURLs starts a local socket server. When a second finicky.exe is
// launched with a URL (or --window), it connects here and hands off the request
// instead of starting a duplicate router.
func listenForURLs() {
	listener, err := net.Listen("unix", pipeAddr())
	if err != nil {
		// Stale socket from a previous crash — remove and retry once.
		os.Remove(pipeAddr())
		listener, err = net.Listen("unix", pipeAddr())
		if err != nil {
			slog.Error("Failed to start URL listener", "error", err)
			return
		}
	}
	defer listener.Close()

	slog.Debug("IPC listener started", "address", pipeAddr())
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("IPC accept error", "error", err)
			continue
		}
		go func() {
			defer conn.Close()
			buf := make([]byte, 8192)
			n, err := conn.Read(buf)
			if err != nil || n == 0 {
				return
			}
			payload := strings.TrimSpace(string(buf[:n]))
			switch {
			case payload == "":
				return
			case payload == showWindowCmd:
				slog.Info("Received show-window request from another instance")
				select {
				case queueWindowOpen <- true:
				default:
				}
			default:
				slog.Info("Received URL from IPC", "url", payload)
				handleURLString(payload, false)
			}
		}()
	}
}

// handOffToPrimary forwards this (secondary) instance's work to the running
// primary over the IPC socket, then the caller exits.
func handOffToPrimary(urlFromArgs string) {
	if urlFromArgs != "" {
		if err := sendToPrimary(urlFromArgs); err != nil {
			slog.Error("Could not forward URL to the running Finicky instance", "error", err)
			// Fall back to the OS default handler ONLY if we are not it —
			// otherwise `start` would relaunch Finicky and loop.
			if def, _ := isDefaultBrowser(); !def {
				cmd := exec.Command("cmd", "/c", "start", "", urlFromArgs)
				cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
				_ = cmd.Run()
			} else {
				slog.Error("Running instance unreachable and Finicky is the default browser; dropping URL to avoid a relaunch loop", "url", urlFromArgs)
			}
		}
		return
	}
	if forceWindowOpen {
		if err := sendToPrimary(showWindowCmd); err != nil {
			slog.Error("Could not ask the running Finicky instance to show its window", "error", err)
		}
	}
}

// sendToPrimary hands a payload (a URL, or the show-window sentinel) to the
// running instance over the local socket. The primary creates its
// single-instance mutex before it begins listening, so a request from a
// near-simultaneous launch can briefly beat the socket into existence; retry
// for ~1s to cover that startup window.
func sendToPrimary(payload string) error {
	var lastErr error
	for i := 0; i < 20; i++ {
		conn, err := net.Dial("unix", pipeAddr())
		if err != nil {
			lastErr = err
			time.Sleep(50 * time.Millisecond)
			continue
		}
		defer conn.Close()
		_ = conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
		_, err = conn.Write([]byte(payload))
		return err
	}
	return lastErr
}

var pipeOnce sync.Once
var pipeAddress string

func pipeAddr() string {
	pipeOnce.Do(func() {
		cacheDir, err := os.UserCacheDir()
		if err != nil {
			pipeAddress = os.TempDir() + `\finicky.sock`
			return
		}
		os.MkdirAll(cacheDir+`\Finicky`, 0755)
		pipeAddress = cacheDir + `\Finicky\finicky.sock`
	})
	return pipeAddress
}
