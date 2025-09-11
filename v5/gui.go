package main

import (
	"github.com/donomii/glim"
	"github.com/go-gl/gl/v2.1/gl"
	"log"
)

var foreColour, backColour *glim.RGBA

func renderEd(w, h int) {
	left := 0
	top := 0

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

		mouseX := 10
		mouseY := 10
		displayText := ed.ActiveBuffer.Data.Text

		log.Println("Render paragraph", len(displayText), "chars")

		glim.RenderPara(ed.ActiveBuffer.Formatter,
			0, 0, 0, 0,
			w, h, w, h,
			int(mouseX)-left, int(mouseY)-top, pic, displayText,
			false, true, true)
		log.Println("Finished render paragraph")
	}
}

func blit(pix []uint8, w, h int) {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	gl.Viewport(0, 0, int32(w), int32(h))
	gl.Ortho(0, 1, 1, 0, 0, -1)

	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	gl.TexImage2D(
		gl.TEXTURE_2D, 0,
		gl.RGBA,
		int32(w), int32(h), 0,
		gl.RGBA,
		gl.UNSIGNED_BYTE, gl.Ptr(pix),
	)

	gl.Enable(gl.TEXTURE_2D)
	gl.Begin(gl.QUADS)
	gl.TexCoord2i(0, 0); gl.Vertex2i(0, 0)
	gl.TexCoord2i(1, 0); gl.Vertex2i(1, 0)
	gl.TexCoord2i(1, 1); gl.Vertex2i(1, 1)
	gl.TexCoord2i(0, 1); gl.Vertex2i(0, 1)
	gl.End()
	gl.Disable(gl.TEXTURE_2D)

	gl.Flush()
	gl.DeleteTextures(1, &texture)
}
