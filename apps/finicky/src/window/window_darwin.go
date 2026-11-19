package window

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa -framework WebKit
#include <stdlib.h>
#include "window.h"
*/
import "C"

func ShowWindow() {
	C.ShowWindow()
}

func CloseWindow() {
	C.CloseWindow()
}

//export GetUIURL
func GetUIURL() *C.char {
	return C.CString(UIURL())
}

//export GetAPIToken
func GetAPIToken() *C.char {
	return C.CString(apiToken)
}
