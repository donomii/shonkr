package main

// #cgo CFLAGS: -g -Wall
// #include <stdlib.h>
// #include "tmt.h"
import "C"

import (
	"fmt"
)

func start_tmt() {

	vt = C.terminal_open()
	C.tmt_resize(vt, 24, 80)

}
func tmt_process_text(vt *C.struct_TMT, text string) {
	C.tmt_write(vt, C.CString(text), 0)
	needsRedraw = true
}

func tmt_resize(width, height uint) {
	C.tmt_resize(vt, C.ulonglong(height), C.ulonglong(width))
}
func tmt_get_screen(vt *C.struct_TMT) string {
	var out string
	scr := C.tmt_screen(vt)
	//fmt.Printf("lines: %v, columns: %v\n", scr.nline, scr.ncol)
	for i := 0; i < int(scr.nline); i++ {
		for j := 0; j < int(scr.ncol); j++ {
			char := fmt.Sprintf("%c", rune(C.terminal_char(vt, C.int(j), C.int(i))))
			out = out + char
		}
		out = out + fmt.Sprintf("\n")
	}
	return out
}

var vt *C.struct_TMT
