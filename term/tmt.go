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
	C.tmt_write(vt, C.CString("\033[1mWelcome to Watterm\033[0m\n"), 0)

}
func tmt_process_text(vt *_Ctype_struct_TMT, text string) {
	C.tmt_write(vt, C.CString(text), 0)
}
func tmt_get_screen(vt *_Ctype_struct_TMT) string {
	var out string
	scr := C.tmt_screen(vt)
	fmt.Printf("lines: %v, columns: %v\n", scr.nline, scr.ncol)
	for i := 0; i < int(scr.nline); i++ {
		for j := 0; j < int(scr.ncol); j++ {
			char := fmt.Sprintf("%c", rune(C.terminal_char(vt, C.int(j), C.int(i))))
			/*
				if char == " " {
					out = out + "X"
				} else {
					out = out + char
				}
			*/
			out = out + char
		}
		out = out + fmt.Sprintf("\n")
	}
	return out
}

var vt *_Ctype_struct_TMT
