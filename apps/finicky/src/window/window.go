package window

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework WebKit
#include <stdlib.h>
#include "window.h"
*/
import "C"
import (
	"finicky/assets"
	"finicky/rules"
	"io/fs"
	"log/slog"
	"net/http"
	"path/filepath"
	"strings"
	"unsafe"
)

var SaveRulesHandler func(rules.RulesFile)

func init() {
	html, err := assets.GetHTML()
	if err != nil {
		slog.Error("Error loading HTML content", "error", err)
		return
	}

	cContent := C.CString(html)
	defer C.free(unsafe.Pointer(cContent))
	C.SetHTMLContent(cContent)

	filesystem := assets.GetFileSystem()
	err = fs.WalkDir(filesystem, "templates", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Base(path) == "index.html" {
			return nil
		}
		content, err := assets.GetFile(filepath.Base(path))
		if err != nil {
			slog.Error("Error loading file", "path", path, "error", err)
			return nil
		}
		cPath := C.CString(filepath.Base(path))
		cContent := C.CString(string(content))
		defer C.free(unsafe.Pointer(cPath))
		defer C.free(unsafe.Pointer(cContent))

		contentType := http.DetectContentType(content)
		if strings.HasPrefix(contentType, "text/") || strings.HasPrefix(contentType, "application/javascript") {
			C.SetFileContent(cPath, cContent)
		} else {
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
}

func CloseWindow() {
	C.CloseWindow()
}

//export GetAPIPort
func GetAPIPort() C.int {
	return C.int(apiPort)
}
