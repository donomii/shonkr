// +build windows

package main

import "github.com/donomii/goof"

func startShell() (chan []byte, chan []byte) {
	shellIn, shellOut, _ := goof.WrapProc("cmd", 100)
	return shellIn, shellOut
}
