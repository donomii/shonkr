package main

import (
	"bytes"
	"os"

	"fmt"
	"io/ioutil"
	"runtime"

	"golang.org/x/image/font/gofont/goregular"

	//"unsafe"

	"time"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"
	"github.com/xlab/closer"

	//"text/scanner"

	"flag"

	"log"

	"github.com/donomii/glim"
)

var shell string
var active bool
var myMenu Menu

type Menu []string

var form *glim.FormatParams
var lasttime float64
var autoSync bool
var ui bool
var repos [][]string
var lastSelect string
var workerChan chan string
var needsRedraw bool

type Option uint8

type State struct {
	bgColor nk.Color
	prop    int32
	opt     Option
}

type UserConfig struct {
	Red, Green, Blue int
}

var winWidth = 900
var winHeight = 900
var ed *GlobalConfig
var config UserConfig
var confFile string

//var stdinQ, stdoutQ, stderrQ,
var shellIn, shellOut chan []byte

// Arrange that main.main runs on main thread.
func init() {
	runtime.LockOSThread()
}

var pic []uint8
var picBytes []byte

func main() {
	runtime.LockOSThread()

	pic = make([]uint8, 3000*3000*4)
	picBytes = make([]byte, 3000*3000*4)
	var doLogs bool
	flag.BoolVar(&doLogs, "debug", false, "Display logging information")
	flag.StringVar(&shell, "shell", "/bin/bash", "The command shell to run")
	flag.Parse()
	if !doLogs {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	start_tmt()
	//time.Sleep(1 * time.Second)
	os.Setenv("TERM", "dumb")
	os.Setenv("LINES", "24")
	os.Setenv("COLUMNS", "80")
	os.Setenv("PS1", ">")

	shellIn, shellOut = startShell(shell)

	foreColour = &glim.RGBA{255, 255, 255, 255}
	backColour = &glim.RGBA{0, 0, 0, 255}

	ed = NewEditor()
	//Create a text formatter.  This controls the appearance of the text, e.g. colour, size, layout
	form = glim.NewFormatter()
	ed.ActiveBuffer.Formatter = form
	SetFont(ed.ActiveBuffer, 12)

	//Nuklear

	if err := glfw.Init(); err != nil {
		closer.Fatalln(err)
	}
	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 2)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	win, err := glfw.CreateWindow(winWidth, winHeight, "ShonkTerm", nil, nil)
	if err != nil {
		closer.Fatalln(err)
	}
	win.MakeContextCurrent()

	win.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

		log.Printf("Got key %c,%v,%v,%v", key, key, mods, action)

		if mods == 2 && action == 1 && key != 341 {
			mask := ^byte(64 + 128)
			log.Printf("mask: %#b", mask)
			val := byte(key)
			log.Printf("val: %#b", val)
			b := mask & val
			log.Printf("byte: %#b", b)
			shellIn <- []byte{b}

		}

		if action == 0 && mods == 0 {
			switch key {
			case 257:
				go func() { shellIn <- []byte("\n") }()
			case 265:
				go func() { shellIn <- []byte("\u001b[A") }()
			case 264:
				go func() { shellIn <- []byte("\u001b[B") }()
			case 263:
				go func() { shellIn <- []byte("\u001b[C") }()
			case 262:
				go func() { shellIn <- []byte("\u001b[D") }()
			case 256:
				go func() { shellIn <- []byte("\u001b") }()
			case 259:
				go func() { shellIn <- []byte{127} }()
			case 258:
				go func() { shellIn <- []byte("\t") }()
			}
		}

	})

	win.SetCharModsCallback(func(w *glfw.Window, char rune, mods glfw.ModifierKey) {

		text := fmt.Sprintf("%c", char)
		fmt.Printf("Text: %v\n", text)
		shellIn <- []byte(text)

	})

	if err := gl.Init(); err != nil {
		closer.Fatalln("opengl: init failed:", err)
	}

	ctx := nk.NkPlatformInit(win, nk.PlatformInstallCallbacks)

	atlas := nk.NewFontAtlas()
	nk.NkFontStashBegin(&atlas)
	sansFont := nk.NkFontAtlasAddFromBytes(atlas, goregular.TTF, 16, nil)
	nk.NkFontStashEnd()
	if sansFont != nil {
		nk.NkStyleSetFont(ctx, sansFont.Handle())
	}

	exitC := make(chan struct{}, 1)
	doneC := make(chan struct{}, 1)
	closer.Bind(func() {
		close(exitC)
		<-doneC
	})

	state := &State{
		bgColor: nk.NkRgba(255, 255, 255, 255),
	}

	go func() {
		for {
			if active {
				log.Println("Waiting for data from stdoutQ")
				data := <-shellOut
				log.Println("Got data", string(data))
				if runtime.GOOS == "windows" {
					data = bytes.Replace(data, []byte("\n"), []byte("\r\n"), -1)
				}
				tmt_process_text(vt, string(data))
				SetBuffer(ed.ActiveBuffer, tmt_get_screen(vt))
				needsRedraw = true
			}
		}
	}()

	fpsTicker := time.NewTicker(time.Second / 30)

	//LoadFileIfNotLoaded(ed, flag.Arg(0))
	SetFont(ed.ActiveBuffer, 12)
	log.Println("Starting main loop")
	needsRedraw = true

	for {
		log.Println("Mainloop!")
		select {
		case <-exitC:
			nk.NkPlatformShutdown()
			glfw.Terminate()
			fpsTicker.Stop()
			close(doneC)
			return
		case <-fpsTicker.C:
			if win.ShouldClose() {
				close(exitC)
				continue
			}
			glfw.PollEvents()
			winWidth, winHeight = win.GetSize()
			if needsRedraw {
				lasttime = glfw.GetTime()
				log.Println("Redraw!")
				gfxMain(win, ctx, state)
				needsRedraw = false
				log.Println("Setting active to true")
				active = true

			} else {
				TARGET_FPS := 10.0
				if glfw.GetTime() < lasttime+1.0/TARGET_FPS {
					time.Sleep(10 * time.Millisecond)
				} else {
					needsRedraw = true
				}
			}

		}
	}

}
