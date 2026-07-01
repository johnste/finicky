//go:build windows

package window

import (
	"encoding/json"
	"finicky/assets"
	"finicky/browser"
	"finicky/rules"
	"finicky/util"
	"finicky/version"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"

	webview2 "github.com/jchv/go-webview2"
)

var (
	messageQueue     []string
	queueMutex       sync.Mutex
	windowReady      bool
	TestUrlHandler   func(string)
	SaveRulesHandler func(rules.RulesFile)

	wv          webview2.WebView
	assetServer net.Listener
)

func SendMessageToWebView(messageType string, message interface{}) {
	jsonMsg := struct {
		Type    string      `json:"type"`
		Message interface{} `json:"message"`
	}{
		Type:    messageType,
		Message: message,
	}
	jsonBytes, err := json.Marshal(jsonMsg)
	if err != nil {
		slog.Error("Error marshaling message", "error", err)
		return
	}

	queueMutex.Lock()
	defer queueMutex.Unlock()

	if windowReady {
		sendMessageToWebViewInternal(string(jsonBytes))
	} else {
		messageQueue = append(messageQueue, string(jsonBytes))
	}
}

// sendMessageToWebViewInternal must be called with queueMutex held.
func sendMessageToWebViewInternal(message string) {
	if wv == nil {
		return
	}
	escaped := strings.ReplaceAll(message, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	js := fmt.Sprintf(`finicky.receiveMessage("%s")`, escaped)
	w := wv
	wv.Dispatch(func() {
		w.Eval(js)
	})
}

// RunWindow creates the WebView2 window and runs its message loop on the
// CURRENT OS thread, blocking until the window is closed.
//
// It MUST be called from the main thread — the goroutine pinned by
// runtime.LockOSThread in main(). Win32 delivers window messages only to the
// thread that created the window, so creating the window on one goroutine and
// pumping its loop on another (as an earlier version did) leaves the UI frozen
// and can crash WebView2 during initialization. Keeping creation and the
// message loop on one locked thread mirrors how the macOS build runs the whole
// Cocoa app on the main thread.
func RunWindow() {
	slog.Debug("Creating window")

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:     true,
		AutoFocus: true,
		WindowOptions: webview2.WindowOptions{
			Title:  "Finicky",
			Width:  860,
			Height: 600,
		},
	})
	if w == nil {
		slog.Error("Failed to create webview2 instance")
		return
	}

	queueMutex.Lock()
	wv = w
	windowReady = false
	queueMutex.Unlock()

	w.SetSize(860, 600, webview2.HintNone)

	// Inject the finicky stub before any page JS runs (mirrors the macOS WKUserScript).
	w.Init(`
		window.finicky = {
			_queue: [],
			receiveMessage: function(msg) { this._queue.push(msg); },
			sendMessage: function(msg) {
				window.__finicky_send(JSON.stringify(msg));
			}
		};
		window.addEventListener("load", function() {
			window.__finicky_ready();
		});
	`)

	// Bind the JS→Go message channel.
	w.Bind("__finicky_send", func(msgJSON string) {
		HandleWebViewMessage(msgJSON)
	})

	// Bind the navigation-complete signal so the message queue drains.
	w.Bind("__finicky_ready", func() {
		NavigationComplete()
	})

	// Start a local HTTP server to serve the embedded Svelte UI assets.
	addr := startAssetServer()
	if addr != "" {
		w.Navigate("http://" + addr + "/index.html")
	} else {
		html, err := assets.GetHTML()
		if err != nil {
			slog.Error("Failed to load HTML", "error", err)
		} else {
			w.SetHtml(html)
		}
	}

	SendBuildInfo()

	// Blocks on the Win32 message loop until the window is closed.
	w.Run()

	// Window closed — reset shared state so a later RunWindow starts clean and
	// queued messages buffer again, and shut down the asset server it was using.
	queueMutex.Lock()
	wv = nil
	windowReady = false
	queueMutex.Unlock()

	if assetServer != nil {
		assetServer.Close()
		assetServer = nil
	}
}

// CloseWindow asks a running window to close, which unblocks RunWindow. Safe to
// call from any goroutine — it marshals onto the UI thread via Dispatch.
func CloseWindow() {
	queueMutex.Lock()
	w := wv
	queueMutex.Unlock()
	if w != nil {
		w.Dispatch(func() {
			w.Terminate()
		})
	}
}

func SendBuildInfo() {
	commitHash, buildDate := version.GetBuildInfo()
	buildInfo := fmt.Sprintf("(%s, built %s)", commitHash, buildDate)
	SendMessageToWebView("buildInfo", buildInfo)
}

func NavigationComplete() {
	queueMutex.Lock()
	windowReady = true
	for _, message := range messageQueue {
		sendMessageToWebViewInternal(message)
	}
	messageQueue = nil
	queueMutex.Unlock()
}

// startAssetServer spins up a localhost HTTP server serving the embedded
// Svelte UI. Returns the listen address (e.g. "127.0.0.1:54321") or "".
func startAssetServer() string {
	mux := http.NewServeMux()

	filesystem := assets.GetFileSystem()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}
		path = strings.TrimPrefix(path, "/")

		content, err := assets.GetFile(path)
		if err != nil {
			// Try reading directly from the embedded FS.
			data, fsErr := fs.ReadFile(filesystem, "templates/"+path)
			if fsErr != nil {
				http.NotFound(w, r)
				return
			}
			content = data
		}

		switch {
		case strings.HasSuffix(path, ".html"):
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
		case strings.HasSuffix(path, ".css"):
			w.Header().Set("Content-Type", "text/css; charset=utf-8")
		case strings.HasSuffix(path, ".js"):
			w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		case strings.HasSuffix(path, ".png"):
			w.Header().Set("Content-Type", "image/png")
		default:
			w.Header().Set("Content-Type", http.DetectContentType(content))
		}
		w.Write(content)
	})

	// Bind to a random available port on loopback only.
	var err error
	assetServer, err = net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		slog.Error("Failed to start asset server", "error", err)
		return ""
	}

	addr := assetServer.Addr().String()
	slog.Debug("Asset server started", "address", addr)

	go http.Serve(assetServer, mux)
	return addr
}

func HandleWebViewMessage(messageStr string) {
	var msg map[string]interface{}
	if err := json.Unmarshal([]byte(messageStr), &msg); err != nil {
		slog.Error("Failed to parse webview message", "error", err)
		return
	}

	messageType, ok := msg["type"].(string)
	if !ok {
		slog.Error("Message missing type field")
		return
	}

	slog.Debug("Received message from webview", "type", messageType)

	switch messageType {
	case "testUrl":
		handleTestUrl(msg)
	case "getRules":
		handleGetRules()
	case "saveRules":
		handleSaveRules(msg)
	case "getInstalledBrowsers":
		handleGetInstalledBrowsers()
	case "getBrowserProfiles":
		handleGetBrowserProfiles(msg)
	default:
		slog.Debug("Unknown message type", "type", messageType)
	}
}

func handleTestUrl(msg map[string]interface{}) {
	url, ok := msg["url"].(string)
	if !ok {
		slog.Error("testUrl message missing url field")
		return
	}
	if TestUrlHandler != nil {
		TestUrlHandler(url)
	} else {
		SendMessageToWebView("testUrlResult", map[string]interface{}{
			"error": "Test handler not initialized",
		})
	}
}

func handleGetRules() {
	rf, err := rules.Load()
	if err != nil {
		slog.Error("Failed to load rules", "error", err)
		SendMessageToWebView("rules", map[string]interface{}{
			"defaultBrowser": "",
			"rules":          []interface{}{},
		})
		return
	}

	path, _ := rules.GetPath()
	var rulesPath string
	if _, statErr := os.Stat(path); statErr == nil {
		rulesPath = path
	}

	type rulesResponse struct {
		rules.RulesFile
		Path string `json:"path,omitempty"`
	}
	SendMessageToWebView("rules", rulesResponse{RulesFile: rf, Path: util.ShortenPath(rulesPath)})
}

func handleSaveRules(msg map[string]interface{}) {
	payload, ok := msg["payload"]
	if !ok {
		slog.Error("saveRules message missing payload field")
		return
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		slog.Error("Failed to marshal saveRules payload", "error", err)
		return
	}

	var rf rules.RulesFile
	if err := json.Unmarshal(payloadBytes, &rf); err != nil {
		slog.Error("Failed to parse saveRules payload", "error", err)
		return
	}

	if err := rules.Save(rf); err != nil {
		slog.Error("Failed to save rules", "error", err)
		SendMessageToWebView("saveRulesError", map[string]interface{}{"error": err.Error()})
		return
	}

	path, _ := rules.GetPath()
	type rulesResponse struct {
		rules.RulesFile
		Path string `json:"path,omitempty"`
	}
	SendMessageToWebView("rules", rulesResponse{RulesFile: rf, Path: util.ShortenPath(path)})

	if SaveRulesHandler != nil {
		SaveRulesHandler(rf)
	}
}

func handleGetInstalledBrowsers() {
	installed := browser.GetInstalledBrowsers()
	SendMessageToWebView("installedBrowsers", installed)
}

func handleGetBrowserProfiles(msg map[string]interface{}) {
	browserName, _ := msg["browser"].(string)
	profiles := browser.GetProfilesForBrowser(browserName)
	SendMessageToWebView("browserProfiles", map[string]interface{}{
		"browser":  browserName,
		"profiles": profiles,
	})
}

