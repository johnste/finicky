// Package picker presents a modal list of destinations and reports which one
// the user chose. It backs the "ask" app type.
package picker

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#include <stdlib.h>
#include "picker.h"
*/
import "C"

import "unsafe"

// Cancelled is returned by Choose when the user dismissed the picker.
const Cancelled = -1

// Choose shows a modal picker listing labels and returns the index of the
// chosen label, or Cancelled if the user dismissed it. tooltip is shown on
// hover and carries the untruncated form of message.
func Choose(title string, message string, tooltip string, labels []string) int {
	if len(labels) == 0 {
		return Cancelled
	}

	cTitle := C.CString(title)
	defer C.free(unsafe.Pointer(cTitle))
	cMessage := C.CString(message)
	defer C.free(unsafe.Pointer(cMessage))
	cTooltip := C.CString(tooltip)
	defer C.free(unsafe.Pointer(cTooltip))

	pointerSize := C.size_t(unsafe.Sizeof(uintptr(0)))
	cLabels := C.malloc(C.size_t(len(labels)) * pointerSize)
	defer C.free(cLabels)

	labelSlice := unsafe.Slice((**C.char)(cLabels), len(labels))
	for i, label := range labels {
		labelSlice[i] = C.CString(label)
	}
	defer func() {
		for i := range labels {
			C.free(unsafe.Pointer(labelSlice[i]))
		}
	}()

	return int(C.showBrowserPicker(cTitle, cMessage, cTooltip, (**C.char)(cLabels), C.int(len(labels))))
}
