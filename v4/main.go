package main

import (
	"flag"
	"fmt"
	"strconv"

	"runtime"
	"time"

	"github.com/donomii/glim"
	"github.com/donomii/menu"

	"io/ioutil"
	"log"
	"os"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.1/glfw"
)

func init() { runtime.LockOSThread() }

var ed *GlobalConfig
var confFile string
var pic []uint8

var pred []string
var input, status string
var selected int
var update bool = true
var form *glim.FormatParams
var edWidth = 1000
var edHeight = 500
var mode = "searching"
var currentMenu *menu.Node

func Seq(min, max int) []int {
	size := max - min + 1
	if size < 1 {
		return []int{}
	}
	a := make([]int, size)
	for i := range a {
		a[i] = min + i
	}
	return a
}

func handleKeys(window *glfw.Window) {
	window.SetKeyCallback(func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

		fmt.Printf("Got key %c,%v,%v,%v\n", key, key, mods, action)
		if action > 0 {
			if key == 256 {
				os.Exit(0)
			}

			if key == 264 {
				selected += 1
				if selected > len(pred)-1 {
					selected = len(pred) - 1
				}
			}

			if mode == "menu" {
				switch key {
				case 301:
					mode = "searching"

				}
			} else {

				switch key {
				case 257:
					ActiveBufferInsert(ed, "\n")
				case 263:
					dispatch("PREVIOUS-CHARACTER", ed)
				case 262:
					dispatch("NEXT-CHARACTER", ed)
				case 265:
					dispatch("PREVIOUS-LINE", ed)
				case 264:
					dispatch("NEXT-LINE", ed)
				case 268:
					dispatch("START-OF-TEXT-ON-LINE", ed)
				case 269:
					dispatch("SEEK-EOL", ed)
				case 259:
					dispatch("DELETE-LEFT", ed)
				case 267:
					dispatch("PAGE-DOWN", ed)
				case 266:
					dispatch("PAGE-UP", ed)
					fmt.Println("Starting pageup with ", edWidth, "x", edHeight)
					PageUp(ed.ActiveBuffer, edWidth, edHeight)
				case 301:
					mode = "menu"
				}
			}
			update = true
		}

	})

	window.SetCharModsCallback(func(w *glfw.Window, char rune, mods glfw.ModifierKey) {
		if mode == "menu" {
			text := fmt.Sprintf("%c", char)
			val, _ := strconv.ParseInt(text, 10, strconv.IntSize)
			fmt.Println("Activating menu option", val)
			item := currentMenu.SubNodes[val]
			fmt.Println("Activating menu option", item.Name)
			f := item.Function
			if f != nil {
				f()
			}

		} else {
			text := fmt.Sprintf("%c", char)
			fmt.Printf("Text input: %v\n", text)
			input = input + text
			ActiveBufferInsert(ed, text)

		}
		update = true

	})
}

func main() {
	var doLogs bool
	flag.BoolVar(&doLogs, "debug", false, "Display logging information")
	flag.Parse()
	filename := ""
	log.Println(flag.Args())
	if len(flag.Args()) > 0 {
		filename = flag.Arg(0)
	}
	log.Println("Loading", filename)

	if !doLogs {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}

	currentMenu = menu.MakeNodeShort("Main Menu", nil)
	item := menu.MakeNodeShort("Go to start", nil)
	item.Function = func() { dispatch("START-OF-FILE", ed); update = true; mode = "searching" }
	menu.AppendNode(currentMenu, item)

	item = menu.MakeNodeShort("Go to end", nil)
	item.Function = func() { dispatch("END-OF-FILE", ed); update = true; mode = "searching" }
	menu.AppendNode(currentMenu, item)

	item = menu.MakeNodeShort("Increase Font", nil)
	item.Function = func() { dispatch("INCREASE-FONT", ed); update = true; mode = "searching" }
	menu.AppendNode(currentMenu, item)

	item = menu.MakeNodeShort("Decrease Font", nil)
	item.Function = func() { dispatch("DECREASE-FONT", ed); update = true; mode = "searching" }
	menu.AppendNode(currentMenu, item)

	item = menu.MakeNodeShort("Vertical Mode", nil)
	item.Function = func() { dispatch("VERTICAL-MODE", ed); update = true; mode = "searching" }
	menu.AppendNode(currentMenu, item)

	item = menu.MakeNodeShort("Horizontal Mode", nil)
	item.Function = func() { dispatch("HORIZONTAL-MODE", ed); update = true; mode = "searching" }
	menu.AppendNode(currentMenu, item)

	item = menu.MakeNodeShort("Save file", nil)
	item.Function = func() { dispatch("SAVE-FILE", ed); update = true; mode = "searching" }
	menu.AppendNode(currentMenu, item)

	item = menu.MakeNodeShort("Switch Buffer", nil)
	item.Function = func() {
		buffMenu := menu.MakeNodeShort("Buffer Menu", nil)
		for i, v := range ed.BufferList {
			ii := i
			item = menu.MakeNodeShort(v.Data.FileName, nil)
			item.Function = func() { ed.ActiveBuffer = ed.BufferList[ii]; update = true }
			menu.AppendNode(buffMenu, item)
		}
		currentMenu = buffMenu
	}
	menu.AppendNode(currentMenu, item)

	log.Println("Init glfw")
	if err := glfw.Init(); err != nil {
		panic("failed to initialize glfw: " + err.Error())
	}
	defer glfw.Terminate()

	log.Println("Setup window")
	monitor := glfw.GetPrimaryMonitor()
	mode := monitor.GetVideoMode()
	edWidth = mode.Width - int(float64(mode.Width)*0.1)
	//edHeight = mode.Height

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	log.Println("Make window")
	window, err := glfw.CreateWindow(edWidth, edHeight, "Menu", nil, nil)
	if err != nil {
		panic(err)
	}
	log.Println("Make context current")
	window.MakeContextCurrent()
	log.Println("Allocate memory")
	pic = make([]uint8, 3000*3000*4)
	ed = NewEditor()
	//Create a text formatter.  This controls the appearance of the text, e.g. colour, size, layout
	form = glim.NewFormatter()
	ed.ActiveBuffer.Formatter = form
	SetFont(ed.ActiveBuffer, 16)
	log.Println("Set up handlers")
	handleKeys(window)

	//This should be SetFramebufferSizeCallback, but that doesn't work, so...
	window.SetSizeCallback(func(w *glfw.Window, width int, height int) {

		edWidth = width
		edHeight = height
		renderEd(edWidth, edHeight)
		blit(pic, edWidth, edHeight)
		window.SwapBuffers()
		update = true
	})

	log.Println("Init gl")
	if err := gl.Init(); err != nil {
		panic(err)
	}
	/*
		go func() {
			lastTime := glfw.GetTime()

			for {
				nowTime := glfw.GetTime()
				if nowTime-lastTime < 10000.0 {

					update = true
					fmt.Println("Forece refresh")
				} else {
					return
				}
			}
		}()
	*/

	lastTime := glfw.GetTime()
	frames := 0

	if filename != "" {
		data, _ := ioutil.ReadFile(filename)
		ActiveBufferInsert(ed, string(data))
		ed.ActiveBuffer.Formatter.Cursor = 0
		ed.ActiveBuffer.Formatter.FirstDrawnCharPos = 0
		ed.ActiveBuffer.Data.FileName = filename
	}

	log.Println("Start rendering")
	for !window.ShouldClose() {
		time.Sleep(35 * time.Millisecond)
		frames++
		nowTime := glfw.GetTime()
		if nowTime-lastTime >= 1.0 {
			//status = fmt.Sprintf("%.3f ms/f  %.0ffps\n", 1000.0/float32(frames), float32(frames))
			frames = 0
			lastTime += 1.0
		}

		if update {
			renderEd(edWidth, edHeight)
			blit(pic, edWidth, edHeight)
			window.SwapBuffers()
			update = false
		}
		glfw.PollEvents()
	}
}

func blit(pix []uint8, w, h int) {
	gl.ClearColor(0.0,1.0,0.0,1.0);
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.MatrixMode(gl.MODELVIEW)
	gl.LoadIdentity()

	gl.Viewport(0, 0, int32(w)*screenScale(), int32(h)*screenScale())
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
	{
		gl.Begin(gl.QUADS)
		{
			gl.TexCoord2i(0, 0)
			gl.Vertex2i(0, 0)

			gl.TexCoord2i(1, 0)
			gl.Vertex2i(1, 0)

			gl.TexCoord2i(1, 1)
			gl.Vertex2i(1, 1)

			gl.TexCoord2i(0, 1)
			gl.Vertex2i(0, 1)
		}
		gl.End()
	}
	gl.Disable(gl.TEXTURE_2D)

	gl.Flush()

	gl.DeleteTextures(1, &texture)
}
