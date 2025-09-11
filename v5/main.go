package main

import (
	"bytes"
	"os"
	"runtime/debug"

	"fmt"
	"io/ioutil"
	"runtime"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"flag"
	"log"

	"net/http"
	_ "net/http/pprof"

	"github.com/donomii/glim"
	"go.uber.org/zap"
)

var shell string
var needsRedraw bool
var active bool

var form *glim.FormatParams
var lasttime float64
var autoSync bool
var ui bool
var repos [][]string
var lastSelect string
var workerChan chan string
var needsRedraw bool
var useAminal bool = false // Disabled for now

type UserConfig struct {
	Red, Green, Blue int
}

var winWidth = 900
var winHeight = 900
var ed *GlobalConfig
var confFile string

// Arrange that main.main runs on main thread.
func init() {
	runtime.LockOSThread()
	debug.SetGCPercent(-1)
}

var pic []uint8
var picBytes []byte

func main() {
	go func() {
		log.Println(http.ListenAndServe("localhost:6060", nil))
	}()
	
	pic = make([]uint8, 3000*3000*4)
	picBytes = make([]byte, 3000*3000*4)
	
	var doLogs bool
	flag.BoolVar(&doLogs, "debug", false, "Display logging information")
	flag.BoolVar(&useAminal, "aminal", false, "Start aminal termal decoder")
	flag.StringVar(&shell, "shell", "/bin/bash", "The command shell to run")
	flag.Parse()
	
	if !doLogs {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	start_tmt()

	os.Setenv("TERM", "xterm")
	os.Setenv("LINES", "24")
	os.Setenv("COLUMNS", "80")
	os.Setenv("PS1", "> ")

	shellIn, shellOut = startShell(shell)

	foreColour = &glim.RGBA{255, 255, 255, 255}
	backColour = &glim.RGBA{0, 0, 0, 255}

	ed = NewEditor()
	form = glim.NewFormatter()
	ed.ActiveBuffer.Formatter = form
	SetFont(ed.ActiveBuffer, 12)

	// Simple GLFW window without nuklear
	if err := glfw.Init(); err != nil {
		log.Fatal(err)
	}
	defer glfw.Terminate()

	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.Resizable, glfw.True)
	
	win, err := glfw.CreateWindow(winWidth, winHeight, "Shonkr Terminal", nil, nil)
	if err != nil {
		log.Fatal(err)
	}
	win.MakeContextCurrent()

	if err := gl.Init(); err != nil {
		log.Fatal("opengl: init failed:", err)
	}

	win.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
		log.Printf("Got key %c,%v,%v,%v", key, key, mods, action)

		if mods == glfw.ModControl && action == glfw.Press && key != glfw.KeyLeftControl && key != glfw.KeyRightControl {
			// Send control characters to shell
			ctrl_char := byte(key - glfw.KeyA + 1)
			if key >= glfw.KeyA && key <= glfw.KeyZ {
				select {
				case shellIn <- []byte{ctrl_char}:
				default:
				}
			}
		}

		if action == glfw.Press || action == glfw.Repeat {
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
		needsRedraw = true
	})

	win.SetCharCallback(func(w *glfw.Window, char rune) {
		text := string(char)
		log.Printf("Text input: %v\n", text)
		select {
		case shellIn <- []byte(text):
		default:
		}
		needsRedraw = true
	})

	win.SetSizeCallback(func(w *glfw.Window, width int, height int) {
		winWidth = width
		winHeight = height
		gl.Viewport(0, 0, int32(width), int32(height))
		needsRedraw = true
	})

	go func() {
		for {
			if active {
				log.Println("Waiting for data from shell")
				select {
				case data := <-shellOut:
					log.Println("Got data", string(data))
					if runtime.GOOS == "windows" {
						data = bytes.Replace(data, []byte("\n"), []byte("\r\n"), -1)
					}
					tmt_process_text(vt, string(data))
					SetBuffer(ed.ActiveBuffer, tmt_get_screen(vt))
					needsRedraw = true
				}
			}
		}
	}()

	SetFont(ed.ActiveBuffer, 12)
	log.Println("Starting main loop")
	needsRedraw = true
	active = true

	for !win.ShouldClose() {
		glfw.PollEvents()
		
		winWidth, winHeight = win.GetSize()
		
		if needsRedraw {
			renderTerminal()
			win.SwapBuffers()
			needsRedraw = false
		}
		
		time.Sleep(16 * time.Millisecond) // ~60 FPS
	}
}

func renderTerminal() {
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	
	// Simple text rendering using glim
	if ed != nil && ed.ActiveBuffer != nil {
		displayText := ed.ActiveBuffer.Data.Text
		
		// Clear the picture buffer
		size := winWidth * winHeight * 4
		if len(pic) < size {
			pic = make([]uint8, size)
		}
		
		// Fill with background color
		for i := 0; i < size; i += 4 {
			pic[i] = 0   // R
			pic[i+1] = 0 // G  
			pic[i+2] = 0 // B
			pic[i+3] = 255 // A
		}
		
		// Render text to the buffer
		form.Colour = &glim.RGBA{255, 255, 255, 255}
		glim.RenderPara(form, 0, 0, 0, 0, winWidth, winHeight, winWidth, winHeight, 10, 10, pic, displayText, false, true, true)
		
		// Render the buffer to screen
		renderBuffer()
	}
}

func renderBuffer() {
	gl.MatrixMode(gl.PROJECTION)
	gl.LoadIdentity()
	gl.Ortho(0, 1, 1, 0, -1, 1)
	
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()
	
	// Create and bind texture
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

//Aminal support functions (simplified stubs)

func getLogger(conf *Config) (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("Failed to create logger: %s", err)
	}
	return logger.Sugar(), nil
}

func getActuallyProvidedFlags() map[string]bool {
	result := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		result[f.Name] = true
	})
	return result
}

func getConfig() *Config {
	return &DefaultConfig
}

func loadConfigFile() *Config {
	return &DefaultConfig
}
