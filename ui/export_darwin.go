package ui

/*
#include <stdlib.h>
*/
import "C"

import (
	"encoding/json"
)

//export goSectors
func goSectors(group, path *C.char) *C.char {
	items := []Item{}
	if cb.Sectors != nil {
		items = cb.Sectors(C.GoString(group), C.GoString(path))
	}
	js, _ := json.Marshal(items)
	return C.CString(string(js)) // freed by the Objective-C caller
}

//export goDrop
func goDrop(id, pathsJSON *C.char) {
	var paths []string
	if err := json.Unmarshal([]byte(C.GoString(pathsJSON)), &paths); err != nil || cb.Drop == nil {
		return
	}
	go cb.Drop(C.GoString(id), paths)
}

//export goHotkey
func goHotkey() {
	if cb.Hotkey != nil {
		go cb.Hotkey()
	}
}
