package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/atotto/clipboard"
	"github.com/donomii/glim"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var winWidth = 900
var winHeight = 900
var needsRedraw = true
var active = false
var ed *GlobalConfig
var form *glim.FormatParams
var pic []uint8
var picBytes []byte

// Terminal backend
var term *TerminalBackend

// Renderer for OpenGL
var rdr Renderer

// Renderer is defined in gui.go
var shellIn chan []byte
var mouseX, mouseY int
var mouseDown bool

var pty *os.File

func init() {
	runtime.LockOSThread()
}

func main() {
	fmt.Println("Starting Shonkr Terminal...")

	pic = make([]uint8, winWidth*winHeight*4)
	picBytes = make([]byte, winWidth*winHeight*4)

	// Initialize editor
	ed = NewEditor()
	form = glim.NewFormatter()
	ed.ActiveBuffer.Formatter = form
	SetFont(ed.ActiveBuffer, 12)

	// Create terminal backend
	term = NewTerminalBackend(80, 24)
	if term == nil {
		log.Fatal("Failed to create terminal backend")
	}

	// Start shell
	shellPath := os.Getenv("SHELL")
	if shellPath == "" {
		shellPath = "/bin/bash"
	}
	shellIn = make(chan []byte, 100)

	os.Setenv("TERM", "xterm-256color")
	os.Setenv("COLORTERM", "truecolor")
	os.Setenv("PS1", "shonkr> ")

	var err error
	pty, err = startShellWithBackend(shellPath, term)
	if err != nil {
		log.Printf("Failed to start shell: %v", err)
	}

	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatal(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	glfw.WindowHint(glfw.Resizable, glfw.True)

	win, err := glfw.CreateWindow(winWidth, winHeight, "Shonkr Terminal", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	win.MakeContextCurrent()

	// Key handling
	win.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		handleKey(key, scancode, action, mods)
	})

	win.SetCharModsCallback(func(w *glfw.Window, char rune, mods glfw.ModifierKey) {
		text := string(char)
		if shellIn != nil {
			shellIn <- []byte(text)
		}
	})

	win.SetCursorPosCallback(func(w *glfw.Window, xpos float64, ypos float64) {
		mouseX = int(xpos)
		mouseY = int(ypos)
	})

	win.SetMouseButtonCallback(func(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
		if button == glfw.MouseButtonLeft {
			mouseDown = (action != glfw.Release)
		}
	})

	if err := gl.Init(); err != nil {
		log.Fatal("opengl: init failed:", err)
	}

	active = true
	needsRedraw = true

	// Main loop
	for !win.ShouldClose() {
		glfw.PollEvents()
		winWidth, winHeight = win.GetSize()
		fbWidth, fbHeight := win.GetFramebufferSize()

		updateTermSize()

		if needsRedraw {
			renderTerminal(fbWidth, fbHeight)
			win.SwapBuffers()
			needsRedraw = false
		}

		time.Sleep(16 * time.Millisecond) // ~60 FPS
	}

	fmt.Println("Shonkr Terminal Closed")
}

var lastWidth, lastHeight int

func updateTermSize() {
	if winWidth == 0 || winHeight == 0 {
		return
	}

	if winWidth == lastWidth && winHeight == lastHeight {
		return
	}
	lastWidth = winWidth
	lastHeight = winHeight

	fontSize := 12.0
	if ed != nil && ed.ActiveBuffer != nil {
		fontSize = ed.ActiveBuffer.Formatter.FontSize
	}

	w, _, lineHeight := glim.GetFontMetrics(fontSize, "M")
	charW := w / 2
	charH := lineHeight

	if charW == 0 || charH == 0 {
		// Avoid divide by zero
		return
	}

	cols := winWidth / charW
	rows := winHeight / charH

	if cols < 1 {
		cols = 1
	}
	if rows < 1 {
		rows = 1
	}

	if term != nil {
		term.SetSize(cols, rows)
	}

	if pty != nil {
		ResizePty(pty, cols, rows)
	}
}

func handleKey(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Press || action == glfw.Repeat {
		switch key {
		case glfw.KeyEnter:
			if shellIn != nil {
				shellIn <- []byte("\n")
			}
		case glfw.KeyBackspace:
			if shellIn != nil {
				shellIn <- []byte{127}
			}
		case glfw.KeyTab:
			if shellIn != nil {
				shellIn <- []byte("\t")
			}
		case glfw.KeyEscape:
			if shellIn != nil {
				shellIn <- []byte("\x1b")
			}
		case glfw.KeyUp:
			if shellIn != nil {
				shellIn <- []byte("\x1b[A")
			}
		case glfw.KeyDown:
			if shellIn != nil {
				shellIn <- []byte("\x1b[B")
			}
		case glfw.KeyLeft:
			if shellIn != nil {
				shellIn <- []byte("\x1b[D")
			}
		case glfw.KeyRight:
			if shellIn != nil {
				shellIn <- []byte("\x1b[C")
			}
		}
	}

	// Handle Ctrl combinations
	if action == glfw.Press {
		// Treat Command (macOS) or Control (others) as "app shortcuts"
		appMod := (mods&glfw.ModSuper) != 0 || (mods&glfw.ModControl) != 0
		if appMod {
			switch key {
			case glfw.KeyA: // Select all
				if ed != nil && ed.ActiveBuffer != nil {
					txt := term.String()
					SetBuffer(ed.ActiveBuffer, txt)
					ed.ActiveBuffer.Formatter.SelectStart = 0
					ed.ActiveBuffer.Formatter.SelectEnd = len(ed.ActiveBuffer.Data.Text) - 1
				}
			case glfw.KeyC: // Copy selection
				if ed != nil && ed.ActiveBuffer != nil {
					if ed.ActiveBuffer.Formatter.SelectStart < 0 || ed.ActiveBuffer.Formatter.SelectEnd <= ed.ActiveBuffer.Formatter.SelectStart {
						// If nothing selected, copy all
						ed.ActiveBuffer.Formatter.SelectStart = 0
						ed.ActiveBuffer.Formatter.SelectEnd = len(ed.ActiveBuffer.Data.Text) - 1
					}
					dispatch("COPY-TO-CLIPBOARD", ed)
				}
			case glfw.KeyV: // Paste to shell
				if shellIn != nil {
					if txt, err := clipboard.ReadAll(); err == nil {
						shellIn <- []byte(txt)
					}
				}
			case glfw.KeyX: // Cut not meaningful; copy
				if ed != nil && ed.ActiveBuffer != nil {
					dispatch("COPY-TO-CLIPBOARD", ed)
				}
			default:
				// Not an app shortcut we handle; fall through to send Ctrl key to shell
			}
			return
		}
		// Raw Ctrl key to shell (e.g., Ctrl+C)
		if (mods & glfw.ModControl) != 0 {
			if key >= glfw.KeyA && key <= glfw.KeyZ {
				ctrl_char := byte(key - glfw.KeyA + 1)
				if shellIn != nil {
					shellIn <- []byte{ctrl_char}
				}
			}
		}
	}
}

func renderTerminal(viewportWidth, viewportHeight int) {
	gl.Viewport(0, 0, int32(viewportWidth), int32(viewportHeight))
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	if ed != nil && ed.ActiveBuffer != nil && term != nil {
		// Clear graphics buffer
		size := winWidth * winHeight * 4
		if len(pic) < size {
			pic = make([]uint8, size)
		}

		// Fill with black background
		for i := 0; i < size; i += 4 {
			pic[i] = 0     // R
			pic[i+1] = 0   // G
			pic[i+2] = 0   // B
			pic[i+3] = 255 // A
		}

		// Get terminal text and sync editor buffer to enable selection/copy
		displayText := term.String()
		if ed != nil && ed.ActiveBuffer != nil {
			SetBuffer(ed.ActiveBuffer, displayText)
		}

		// Render text
		if len(displayText) > 0 {
			form.Colour = &glim.RGBA{255, 255, 255, 255}
			// Enable selection outlines and pass real mouse position (relative to text origin at 10,10)
			ed.ActiveBuffer.Formatter.Outline = true
			mx := mouseX - 10
			my := mouseY - 10
			if mx < 0 {
				mx = 0
			}
			if my < 0 {
				my = 0
			}
			glim.RenderPara(form, 0, 0, 0, 0, winWidth, winHeight, winWidth, winHeight, mx, my, pic, displayText, false, true, true)
		}

		// Display the rendered buffer
		renderBuffer()
	}
}

func renderBuffer() {
	// Use shared core-profile renderer
	if err := rdr.Init(); err != nil {
		log.Printf("renderer init failed: %v", err)
		return
	}
	rdr.UpdateTexture(pic, winWidth, winHeight)
	rdr.Draw()
}
