package window

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework WebKit
#include <stdlib.h>
#include "window.h"
*/
import "C"
import (
	"encoding/json"
	"finicky/assets"
	"finicky/version"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"unsafe"
)

var (
	messageQueue []string
	queueMutex   sync.Mutex
	windowReady  bool
	TestUrlHandler func(string)
)

//export WindowIsReady
func WindowIsReady() {
	queueMutex.Lock()
	windowReady = true
	// Process any queued messages
	for _, message := range messageQueue {
		sendMessageToWebViewInternal(message)
	}
	messageQueue = nil
	queueMutex.Unlock()
}

func sendMessageToWebViewInternal(message string) {
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage))
	C.SendMessageToWebView(cMessage)
}

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

func init() {
	// Load HTML content
	html, err := assets.GetHTML()
	if err != nil {
		slog.Error("Error loading HTML content", "error", err)
		return
	}

	// Set HTML content
	cContent := C.CString(html)
	defer C.free(unsafe.Pointer(cContent))
	C.SetHTMLContent(cContent)

	// Get the filesystem and walk through all files in templates directory
	filesystem := assets.GetFileSystem()
	err = fs.WalkDir(filesystem, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Skip directories and index.html (already handled by GetHTML)
		if d.IsDir() || filepath.Base(path) == "index.html" {
			return nil
		}

		// Get the file content
		content, err := assets.GetFile(filepath.Base(path))
		if err != nil {
			slog.Error("Error loading file", "path", path, "error", err)
			return nil
		}

		cPath := C.CString(filepath.Base(path))
		cContent := C.CString(string(content))
		defer C.free(unsafe.Pointer(cPath))
		defer C.free(unsafe.Pointer(cContent))

		// Detect content type
		contentType := http.DetectContentType(content)
		if strings.HasPrefix(contentType, "text/") || strings.HasPrefix(contentType, "application/javascript") {
			C.SetFileContent(cPath, cContent)
		} else {
			// Handle binary files
			C.SetFileContentWithLength(cPath, cContent, C.size_t(len(content)))
		}
		return nil
	})

	if err != nil {
		slog.Error("Error walking templates directory", "error", err)
	}
}

func ShowWindow() {
	C.ShowWindow()
	SendBuildInfo()
}

func CloseWindow() {
	C.CloseWindow()
}

func SendBuildInfo() {
	commitHash, buildDate := version.GetBuildInfo()
	buildInfo := fmt.Sprintf("(%s, built %s)", commitHash, buildDate)
	SendMessageToWebView("buildInfo", buildInfo)
}

//export HandleWebViewMessage
func HandleWebViewMessage(messagePtr *C.char) {
	messageStr := C.GoString(messagePtr)

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

	slog.Debug("Forwarding test URL request", "url", url)

	if TestUrlHandler != nil {
		TestUrlHandler(url)
	} else {
		slog.Error("TestUrlHandler not set")
		SendMessageToWebView("testUrlResult", map[string]interface{}{
			"error": "Test handler not initialized",
		})
	}
}
