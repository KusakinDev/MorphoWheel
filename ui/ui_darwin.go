// Package ui is the native macOS layer: menu bar item, global hotkey and the
// circular drop menu. The Objective-C side lives in bridge.m.
package ui

/*
#cgo CFLAGS: -x objective-c -fobjc-arc -mmacosx-version-min=12.0
#cgo LDFLAGS: -framework Cocoa -framework Carbon
#include <stdlib.h>
#include "bridge.h"
*/
import "C"

import (
	"encoding/json"
	"unsafe"
)

// Item is one ring sector.
type Item struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

// Callbacks connect the native layer to application logic.
type Callbacks struct {
	// Sectors lists the ring items for a group ("convert", "tools",
	// "actions") given the first dragged file.
	Sectors func(group, path string) []Item
	// Drop runs the chosen operation on the files.
	Drop func(id string, paths []string)
	// Hotkey is invoked on Control+Option+Space or the menu bar entry.
	Hotkey func()
}

var cb Callbacks

// Run blocks on the Cocoa main loop. It must be called from main().
func Run(c Callbacks) {
	cb = c
	C.MorphoWheelRun()
}

// ShowRing displays the ring at the cursor for the given files.
func ShowRing(group string, paths []string) {
	js, _ := json.Marshal(paths)
	g, p := C.CString(group), C.CString(string(js))
	defer C.free(unsafe.Pointer(g))
	defer C.free(unsafe.Pointer(p))
	C.MorphoWheelShowRing(g, p)
}

// Notify shows a banner notification.
func Notify(title, body string) {
	t, b := C.CString(title), C.CString(body)
	defer C.free(unsafe.Pointer(t))
	defer C.free(unsafe.Pointer(b))
	C.MorphoWheelNotify(t, b)
}
