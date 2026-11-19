package window

/*
#cgo pkg-config: gtk+-3.0 webkit2gtk-4.1
#include <stdlib.h>
#include "window_linux.h"
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

var (
	// Headless runs without a window: ShowWindow prints the UI's address
	// instead, for use in a regular browser or by tests.
	Headless bool
	// OnClose is called when the window is closed.
	OnClose func()
	// OnOpenURL is called with links clicked in the UI.
	OnOpenURL func(url string)
)

// Run starts the GTK main loop on the calling thread, which must be the main
// thread, and never returns. In headless mode it just blocks.
func Run() {
	if Headless {
		select {}
	}
	C.FinickyInitWindowing()
	C.FinickyRunMainLoop()
}

func ShowWindow() {
	if Headless {
		fmt.Fprintf(os.Stderr, "Finicky UI: %s?apiToken=%s\n", UIURL(), apiToken)
		return
	}
	cURL := C.CString(UIURL())
	defer C.free(unsafe.Pointer(cURL))
	cScript := C.CString(fmt.Sprintf("window.__FINICKY_API_TOKEN__ = '%s';", apiToken))
	defer C.free(unsafe.Pointer(cScript))
	C.FinickyShowWindow(cURL, cScript)
}

func CloseWindow() {
	if !Headless {
		C.FinickyCloseWindow()
	}
}

//export finickyWindowClosed
func finickyWindowClosed() {
	if OnClose != nil {
		go OnClose()
	}
}

//export finickyOpenURL
func finickyOpenURL(url *C.char) {
	if OnOpenURL != nil {
		go OnOpenURL(C.GoString(url))
	}
}
