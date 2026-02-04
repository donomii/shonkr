package main

import "github.com/go-gl/glfw/v3.2/glfw"

var clipboardWindow *glfw.Window

func SetClipboardWindow(w *glfw.Window) {
	clipboardWindow = w
}

func clipboardRead() string {
	if clipboardWindow == nil {
		return ""
	}
	txt, err := clipboardWindow.GetClipboardString()
	if err != nil {
		return ""
	}
	return txt
}

func clipboardWrite(text string) {
	if clipboardWindow == nil {
		return
	}
	clipboardWindow.SetClipboardString(text)
}
