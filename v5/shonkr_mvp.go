package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"regexp"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/donomii/glim"
)

// Global variables
var winWidth = 900
var winHeight = 900
var needsRedraw = true
var active = false
var ed *GlobalConfig
var form *glim.FormatParams
var pic []uint8
var simpleTerminal *SimpleVT

// Terminal emulator structure
type SimpleVT struct {
	width, height   int
	cursorX, cursorY int
	screen          [][]rune
	buffer          strings.Builder
}

func init() {
	runtime.LockOSThread()
}

func main() {
	fmt.Println("Starting Shonkr Terminal MVP...")

	// Initialize terminal
	start_tmt()
	
	// Initialize editor
	ed = NewEditor()
	form = glim.NewFormatter()
	ed.ActiveBuffer.Formatter = form
	SetFont(ed.ActiveBuffer, 14)

	// Initialize graphics buffer
	pic = make([]uint8, winWidth*winHeight*4)

	// Start shell
	shellIn, shellOut := startShell("/bin/bash")
	
	// Initialize GLFW
	if err := glfw.Init(); err != nil {
		log.Fatal("Failed to initialize GLFW:", err)
	}
	defer glfw.Terminate()

	// Create window
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.Resizable, glfw.True)
	
	window, err := glfw.CreateWindow(winWidth, winHeight, "Shonkr Terminal MVP", nil, nil)
	if err != nil {
		log.Fatal("Failed to create window:", err)
	}
	
	window.MakeContextCurrent()

	// Initialize OpenGL
	if err := gl.Init(); err != nil {
		log.Fatal("Failed to initialize OpenGL:", err)
	}

	fmt.Printf("OpenGL Version: %s\n", gl.GoStr(gl.GetString(gl.VERSION)))

	// Set up key handlers
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		if action == glfw.Press || action == glfw.Repeat {
			handleKey(key, mods, shellIn)
		}
	})

	window.SetCharCallback(func(w *glfw.Window, char rune) {
		text := string(char)
		select {
		case shellIn <- []byte(text):
		default:
		}
	})

	window.SetSizeCallback(func(w *glfw.Window, width int, height int) {
		winWidth = width
		winHeight = height
		gl.Viewport(0, 0, int32(width), int32(height))
		// Resize graphics buffer
		pic = make([]uint8, winWidth*winHeight*4)
		needsRedraw = true
	})

	// Handle shell output
	go func() {
		for data := range shellOut {
			log.Printf("Shell output: %s", string(data))
			tmt_process_text(nil, string(data))
			updateDisplay()
		}
	}()

	active = true
	log.Println("Starting main loop...")

	// Main render loop
	for !window.ShouldClose() {
		glfw.PollEvents()
		
		if needsRedraw {
			renderTerminal()
			window.SwapBuffers()
			needsRedraw = false
		}
		
		time.Sleep(16 * time.Millisecond) // ~60 FPS
	}

	fmt.Println("Shonkr Terminal MVP Closed")
}

func handleKey(key glfw.Key, mods glfw.ModifierKey, shellIn chan []byte) {
	switch key {
	case glfw.KeyEnter:
		select {
		case shellIn <- []byte("\n"):
		default:
		}
	case glfw.KeyBackspace:
		select {
		case shellIn <- []byte("\b"):
		default:
		}
	case glfw.KeyTab:
		select {
		case shellIn <- []byte("\t"):
		default:
		}
	case glfw.KeyEscape:
		select {
		case shellIn <- []byte("\x1b"):
		default:
		}
	case glfw.KeyUp:
		select {
		case shellIn <- []byte("\x1b[A"):
		default:
		}
	case glfw.KeyDown:
		select {
		case shellIn <- []byte("\x1b[B"):
		default:
		}
	case glfw.KeyLeft:
		select {
		case shellIn <- []byte("\x1b[D"):
		default:
		}
	case glfw.KeyRight:
		select {
		case shellIn <- []byte("\x1b[C"):
		default:
		}
	}
}

func renderTerminal() {
	// Clear background
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	
	if ed != nil && ed.ActiveBuffer != nil {
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
		
		// Render text
		displayText := ed.ActiveBuffer.Data.Text
		if len(displayText) > 0 {
			form.Colour = &glim.RGBA{255, 255, 255, 255}
			glim.RenderPara(form, 0, 0, 0, 0, winWidth, winHeight, winWidth, winHeight, 10, 10, pic, displayText, false, true, true)
		}
		
		// Display the rendered buffer
		renderBuffer()
	}
}

func renderBuffer() {
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, 1, 1, 0, -1, 1)
	
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	
	// Create texture
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(winWidth), int32(winHeight), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(pic))
	
	gl.Enable(gl.TEXTURE_2D)
	gl.Begin(gl.QUADS)
	gl.TexCoord2f(0, 0); gl.Vertex2f(0, 0)
	gl.TexCoord2f(1, 0); gl.Vertex2f(1, 0)
	gl.TexCoord2f(1, 1); gl.Vertex2f(1, 1)
	gl.TexCoord2f(0, 1); gl.Vertex2f(0, 1)
	gl.End()
	gl.Disable(gl.TEXTURE_2D)
	
	gl.DeleteTextures(1, &texture)
}

func updateDisplay() {
	if simpleTerminal != nil {
		screenText := tmt_get_screen(nil)
		if ed != nil && ed.ActiveBuffer != nil {
			SetBuffer(ed.ActiveBuffer, screenText)
			needsRedraw = true
		}
	}
}

// Simple shell launcher (no CGO)
func startShell(shellPath string) (chan []byte, chan []byte) {
	log.Printf("Starting shell: %s", shellPath)
	
	shellIn := make(chan []byte, 100)
	shellOut := make(chan []byte, 100)
	
	cmd := exec.Command(shellPath)
	cmd.Env = os.Environ()
	
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("Failed to get stdin pipe: %v", err)
		return shellIn, shellOut
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to get stdout pipe: %v", err)
		return shellIn, shellOut
	}
	
	err = cmd.Start()
	if err != nil {
		log.Printf("Failed to start shell: %v", err)
		return shellIn, shellOut
	}
	
	// Handle input to shell
	go func() {
		defer stdin.Close()
		for data := range shellIn {
			_, err := stdin.Write(data)
			if err != nil {
				log.Printf("Error writing to shell: %v", err)
				return
			}
		}
	}()
	
	// Handle output from shell
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text() + "\n"
			select {
			case shellOut <- []byte(line):
			default:
				log.Println("Shell output buffer full")
			}
		}
	}()
	
	return shellIn, shellOut
}

// Simple terminal emulator (no CGO)
func start_tmt() {
	simpleTerminal = &SimpleVT{
		width:   80,
		height:  24,
		cursorX: 0,
		cursorY: 0,
		screen:  make([][]rune, 24),
	}
	
	for i := range simpleTerminal.screen {
		simpleTerminal.screen[i] = make([]rune, 80)
		for j := range simpleTerminal.screen[i] {
			simpleTerminal.screen[i][j] = ' '
		}
	}
}

func tmt_process_text(vt interface{}, text string) {
	if simpleTerminal == nil {
		return
	}
	
	// Strip ANSI codes for simple rendering
	cleanText := stripANSI(text)
	
	for _, char := range cleanText {
		switch char {
		case '\n':
			simpleTerminal.cursorY++
			simpleTerminal.cursorX = 0
			if simpleTerminal.cursorY >= simpleTerminal.height {
				// Scroll up
				copy(simpleTerminal.screen[0:], simpleTerminal.screen[1:])
				simpleTerminal.screen[simpleTerminal.height-1] = make([]rune, simpleTerminal.width)
				for j := range simpleTerminal.screen[simpleTerminal.height-1] {
					simpleTerminal.screen[simpleTerminal.height-1][j] = ' '
				}
				simpleTerminal.cursorY = simpleTerminal.height - 1
			}
		case '\r':
			simpleTerminal.cursorX = 0
		case '\b':
			if simpleTerminal.cursorX > 0 {
				simpleTerminal.cursorX--
			}
		default:
			if simpleTerminal.cursorX < simpleTerminal.width && simpleTerminal.cursorY < simpleTerminal.height {
				simpleTerminal.screen[simpleTerminal.cursorY][simpleTerminal.cursorX] = char
				simpleTerminal.cursorX++
				if simpleTerminal.cursorX >= simpleTerminal.width {
					simpleTerminal.cursorX = 0
					simpleTerminal.cursorY++
					if simpleTerminal.cursorY >= simpleTerminal.height {
						simpleTerminal.cursorY = simpleTerminal.height - 1
					}
				}
			}
		}
	}
}

func tmt_get_screen(vt interface{}) string {
	if simpleTerminal == nil {
		return ""
	}
	
	var result strings.Builder
	for i, row := range simpleTerminal.screen {
		for _, char := range row {
			result.WriteRune(char)
		}
		if i < len(simpleTerminal.screen)-1 {
			result.WriteRune('\n')
		}
	}
	return result.String()
}

func stripANSI(input string) string {
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)
	return ansiRegex.ReplaceAllString(input, "")
}
