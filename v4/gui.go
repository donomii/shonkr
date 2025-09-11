// gui.go
package main

import (
	"runtime/debug"

	"github.com/donomii/glim"

	"fmt"

	"log"
)

var foreColour, backColour *glim.RGBA

func renderEd(w, h int) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in renderEd", r)
			debug.PrintStack()
		}
	}()
	left := 0
	top := 0

	if ed != nil {
		log.Println("Starting editor draw")

		size := w * h * 4
		log.Println("Clearing", size, "bytes(", w, "x", h, ")")
		backColour = &glim.RGBA{0, 0, 0, 0}
		foreColour = &glim.RGBA{255, 255, 255, 255}
		for i := 0; i < size; i = i + 4 {
			pic[i] = ((*backColour)[0])
			pic[i+1] = ((*backColour)[1])
			pic[i+2] = ((*backColour)[2])
			pic[i+3] = ((*backColour)[3])
		}

		form = ed.ActiveBuffer.Formatter
		form.Colour = foreColour
		form.Outline = true

		if mode == "menu" {
			menuForm := glim.NewFormatter()
			displayText := currentMenu.Name + "\nxxxxxxxxxxxxxx\n"
			for i, v := range currentMenu.SubNodes {
				displayText = fmt.Sprintf("%v\n%v) %v", displayText, i, v.Name)
			}

			ed.ActiveBuffer.Formatter.Cursor = 0
			ed.ActiveBuffer.Formatter.FirstDrawnCharPos = 0
			menuForm.FontSize = 32
			menuForm.Colour = foreColour
			menuForm.Outline = true
			glim.RenderPara(menuForm,
				0, 0, 0, 0,
				w, h, w, h,
				10, 10, pic, displayText,
				false, true, true)
		} else {
			mouseX := 10
			mouseY := 10
			displayText := ed.ActiveBuffer.Data.Text

			log.Println("Render paragraph", string(displayText))

			//ed.ActiveBuffer.Formatter.FontSize = 32
			glim.RenderPara(ed.ActiveBuffer.Formatter,
				0, 0, 0, 0,
				w, h, w, h,
				int(mouseX)-left, int(mouseY)-top, pic, displayText,
				false, true, true)
			log.Println("Finished render paragraph")
		}
	}
}
