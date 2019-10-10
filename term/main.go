package main

import (
	"os"

	//"os"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/donomii/goof"

	"github.com/BurntSushi/toml"

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

var form *glim.FormatParams

var autoSync bool
var ui bool
var repos [][]string
var lastSelect string
var workerChan chan string

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

func main() {
	os.Setenv("TERM", "dumb")
	os.Setenv("LINES", "24")
	os.Setenv("COLUMNS", "80")
	os.Setenv("PS1", ">")
	shellIn, shellOut = startShell()
	shellIn <- []byte("ls\n")

	start_tmt()
	confFile = goof.ConfigFilePath(".shonkr.json")
	log.Println("Loading config from:", confFile)
	configBytes, conferr := ioutil.ReadFile(confFile)
	if conferr != nil {
		log.Println("Writing fresh config to:", confFile)
		ioutil.WriteFile(confFile, []byte("test"), 0644)
		configBytes, conferr = ioutil.ReadFile(confFile)
	}

	toml.Decode(string(configBytes), &config)
	flag.BoolVar(&autoSync, "auto-sync", false, "Automatically push then pull on clean repositories")
	flag.BoolVar(&ui, "ui", false, "Experimental graphical user interface")
	flag.Parse()

	fmt.Println("File", flag.Arg(0))

	ed = NewEditor()
	//Create a text formatter
	form = glim.NewFormatter()
	SetFont(ed.ActiveBuffer, 8)

	jsonerr := json.Unmarshal([]byte(menuData), &myMenu)
	if jsonerr != nil {
		fmt.Println(jsonerr)
	}

	//Nuklear

	DirFiles = goof.Ls(".")

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
		log.Printf("Got key %v,%v,%v", key, mods, action)

		if mods == 2 && action == 1 && key != 341 {
			mask := ^byte(64 + 128)
			log.Printf("mask: %#b", mask)
			val := byte(key)
			log.Printf("val: %#b", val)
			b := mask & val
			log.Printf("byte: %#b", b)
			shellIn <- []byte{b}

		}

	})

	width, height := win.GetSize()
	log.Printf("glfw: created window %vx%v", width, height)

	if err := gl.Init(); err != nil {
		closer.Fatalln("opengl: init failed:", err)
	}
	gl.Viewport(0, 0, int32(width-1), int32(height-1))

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
			log.Println("Waiting for data from stdoutQ")
			data := <-shellOut
			data = bytes.Replace(data, []byte("\n"), []byte("\r\n"), -1)

			tmt_process_text(vt, string(data))
			SetBuffer(ed.ActiveBuffer, tmt_get_screen(vt))

		}
	}()

	fpsTicker := time.NewTicker(time.Second / 30)

	LoadFileIfNotLoaded(ed, flag.Arg(0))
	SetFont(ed.ActiveBuffer, 8)
	for {
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
			//log.Printf("glfw: created window %dx%d", width, height)
			gfxMain(win, ctx, state)
		}
	}

}
