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
	"unsafe"
)

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

	cMessage := C.CString(string(jsonBytes))
	defer C.free(unsafe.Pointer(cMessage))
	C.SendMessageToWebView(cMessage)
}

func SendBuildInfo() {
	commitHash, buildDate := version.GetBuildInfo()
	buildInfo := fmt.Sprintf("(%s, built %s)", commitHash, buildDate)
	SendMessageToWebView("buildInfo", buildInfo)
}