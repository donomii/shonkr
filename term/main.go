package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/atotto/clipboard"
	"github.com/donomii/glim"
	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
)

var winWidth = 900
var winHeight = 900
var needsRedraw int32 = 1 // 1 for true, 0 for false
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
		if char == '\n' || char == '\r' {
			return
		}
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
	atomic.StoreInt32(&needsRedraw, 1)

	// Main loop
	for !win.ShouldClose() {
		glfw.PollEvents()
		winWidth, winHeight = win.GetSize()
		fbWidth, fbHeight := win.GetFramebufferSize()

		updateTermSize()

		// Atomically check and reset flag
		if atomic.SwapInt32(&needsRedraw, 0) == 1 {
			renderTerminal(fbWidth, fbHeight)
			win.SwapBuffers()
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
				shellIn <- []byte("\r")
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

		// Get terminal screen state
		screen := term.GetScreen()
		cursorX, cursorY := term.GetCursor()
		// width, _ := term.GetSize()

		// Build tokens and calculate cursor index
		var tokens []glim.Token
		var cursorIndex int = -1

		// We iterate line by line.
		// term.GetScreen returns lines.
		// We will append \n to each line to match glim's expected flow

		currentIndex := 0
		for y, line := range screen {
			for x, cell := range line.Cells {
				// Check if this is the cursor position
				if x == cursorX && y == cursorY {
					cursorIndex = currentIndex
				}

				// Convert simple text. TODO: Handle colors
				// For now, mapping everything to white, but setting up structure
				// We can handle colors later or now... let's try basic colors now

				fg := getColor(cell.FG)
				// bg := getColor(cell.BG) // TODO: background support in glim tokens?

				t := glim.Token{
					Text: string(cell.Char),
					Style: glim.Style{
						ForegroundColour: fg,
					},
				}
				tokens = append(tokens, t)
				currentIndex++
			}

			// Add newline
			tokens = append(tokens, glim.Token{Text: "\n", Style: glim.Style{ForegroundColour: &glim.RGBA{255, 255, 255, 255}}})
			currentIndex++
		}

		// If cursor is at the end of input
		if cursorIndex == -1 && cursorX == 0 && cursorY >= len(screen) {
			cursorIndex = currentIndex
		}

		// Configure formatter
		if ed.ActiveBuffer.Formatter == nil {
			ed.ActiveBuffer.Formatter = glim.NewFormatter()
		}
		form := ed.ActiveBuffer.Formatter
		form.Colour = &glim.RGBA{255, 255, 255, 255}
		form.Cursor = cursorIndex
		form.CursorColour = &glim.RGBA{200, 200, 200, 255} // Visible cursor

		// Render
		glim.RenderTokenPara(form, 0, 0, 0, 0, winWidth, winHeight, winWidth, winHeight, mouseX, mouseY, pic, tokens, false, true, true)

		// Display the rendered buffer
		renderBuffer()
	}
}

// Basic ANSI 256 color mapping (simplified)
func getColor(colorIndex int) *glim.RGBA {
	// Default foreground (usually code 7 or similar in some maps, but vt10x stores raw index)
	// -1 or specific values might be default. assume white for standard
	if colorIndex < 0 || colorIndex > 255 {
		return &glim.RGBA{255, 255, 255, 255}
	}

	// Standard ANSI colors (0-15)
	// 0: Black, 1: Red...
	// This is a rough map, refining it would be good
	switch colorIndex {
	case 0:
		return &glim.RGBA{0, 0, 0, 255}
	case 1:
		return &glim.RGBA{170, 0, 0, 255}
	case 2:
		return &glim.RGBA{0, 170, 0, 255}
	case 3:
		return &glim.RGBA{170, 85, 0, 255}
	case 4:
		return &glim.RGBA{0, 0, 170, 255}
	case 5:
		return &glim.RGBA{170, 0, 170, 255}
	case 6:
		return &glim.RGBA{0, 170, 170, 255}
	case 7:
		return &glim.RGBA{170, 170, 170, 255}
	case 8:
		return &glim.RGBA{85, 85, 85, 255}
	case 9:
		return &glim.RGBA{255, 85, 85, 255}
	case 10:
		return &glim.RGBA{85, 255, 85, 255}
	case 11:
		return &glim.RGBA{255, 255, 85, 255}
	case 12:
		return &glim.RGBA{85, 85, 255, 255}
	case 13:
		return &glim.RGBA{255, 85, 255, 255}
	case 14:
		return &glim.RGBA{85, 255, 255, 255}
	case 15:
		return &glim.RGBA{255, 255, 255, 255}
	}

	// For 256 colors, we'd need a formula or table.
	// For now, return White for anything else to ensure visibility
	return &glim.RGBA{255, 255, 255, 255}
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
