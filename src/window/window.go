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
	"log"
	"strings"
	"unsafe"
)

func init() {
	// Load HTML content
	html, err := assets.GetHTML()
	if err != nil {
		log.Printf("Error loading HTML content: %v", err)
		return
	}

	// Set HTML content
	cContent := C.CString(html)
	defer C.free(unsafe.Pointer(cContent))
	C.SetHTMLContent(cContent)

	// Load and set static files
	files := []string{"styles.css", "app.js", "finicky-icon.png"}
	for _, file := range files {
		content, err := assets.GetFile(file)
		if err != nil {
			log.Printf("Error loading file %s: %v", file, err)
			continue
		}

		cPath := C.CString(file)
		cContent := C.CString(string(content))
		defer C.free(unsafe.Pointer(cPath))
		defer C.free(unsafe.Pointer(cContent))

		if strings.HasSuffix(file, ".png") {
			C.SetFileContentWithLength(cPath, cContent, C.size_t(len(content)))
		} else {
			C.SetFileContent(cPath, cContent)
		}
	}
}

func ShowWindow() {
	C.ShowWindow()
	SendBuildInfo()
}

func CloseWindow() {
	C.CloseWindow()
}

func SendMessageToWebView(messageType string, message string) {
	jsonMsg := struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	}{
		Type:    messageType,
		Message: message,
	}
	jsonBytes, err := json.Marshal(jsonMsg)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}

	message = string(jsonBytes)
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage))
	C.SendMessageToWebView(cMessage)
}

func SendBuildInfo() {
	commitHash, buildDate := version.GetBuildInfo()
	buildInfo := fmt.Sprintf("(%s, built %s)", commitHash, buildDate)
	SendMessageToWebView("buildInfo", buildInfo)
}