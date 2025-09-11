package main

import (
	"github.com/donomii/glim"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"log"
)

var foreColour, backColour *glim.RGBA

// Simple replacement for nuklear GUI
type State struct {
	bgColor interface{}
	prop    int32
	opt     interface{}
}

var rdr Renderer

func gfxMain(win *glfw.Window, ctx interface{}, state *State) {
	log.Println("Starting simple gfx")
	width, height := win.GetSize()
	log.Printf("glfw: window %vx%v", width, height)

	// Clear screen
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	// Render terminal content
	if ed != nil {
		if err := rdr.Init(); err != nil {
			log.Printf("renderer init failed: %v", err)
			return
		}
		renderEd(width, height)
		rdr.UpdateTexture(pic, width, height)
		rdr.Draw()
	}

	win.SwapBuffers()
	log.Println("Finished simple gfx")
}

func renderEd(w, h int) {
	if ed != nil {
		log.Println("Starting editor draw")
		size := w * h * 4
		log.Println("Clearing", size, "bytes(", w, "x", h, ")")

		backColour = &glim.RGBA{0, 0, 0, 255}
		foreColour = &glim.RGBA{255, 255, 255, 255}

		// Clear buffer
		for i := 0; i < size; i = i + 4 {
			pic[i] = ((*backColour)[0])
			pic[i+1] = ((*backColour)[1])
			pic[i+2] = ((*backColour)[2])
			pic[i+3] = ((*backColour)[3])
		}

		form = ed.ActiveBuffer.Formatter
		form.Colour = foreColour
		displayText := tmt_get_screen(nil)
		log.Println("Render paragraph", len(displayText), "chars")

		glim.RenderPara(ed.ActiveBuffer.Formatter,
			0, 0, 0, 0,
			w, h, w, h,
			10, 10, pic, displayText,
			false, true, true)
		log.Println("Finished render paragraph")
	}
}

// Blitting is handled by Renderer in Core profile
