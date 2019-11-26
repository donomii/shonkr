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

	"path/filepath"

	"os/user"

	"github.com/donomii/glim"
	"github.com/liamg/aminal/config"
	"github.com/liamg/aminal/platform"
	"github.com/liamg/aminal/terminal"
	"github.com/riywo/loginshell"
	"go.uber.org/zap"
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
var confFile string

//var stdinQ, stdoutQ, stderrQ,
var shellIn, shellOut chan []byte

// Arrange that main.main runs on main thread.
func init() {
	runtime.LockOSThread()
}

var pic []uint8
var picBytes []byte
var aminal *terminal.Terminal

func startAminal() *terminal.Terminal {
	conf := getConfig()
	logger, err := getLogger(conf)
	if err != nil {
		fmt.Printf("Failed to create logger: %s\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	logger.Infof("Allocating pty...")

	pty, err := platform.NewPty(80, 25)
	if err != nil {
		logger.Fatalf("Failed to allocate pty: %s", err)
	}

	shellStr := conf.Shell
	if shellStr == "" {
		loginShell, err := loginshell.Shell()
		if err != nil {
			logger.Fatalf("Failed to ascertain your shell: %s", err)
		}
		shellStr = loginShell
	}

	os.Setenv("TERM", "xterm-256color") // controversial! easier than installing terminfo everywhere, but obviously going to be slightly different to xterm functionality, so we'll see...
	os.Setenv("COLORTERM", "truecolor")

	guestProcess, err := pty.CreateGuestProcess(shellStr)
	if err != nil {
		pty.Close()
		logger.Fatalf("Failed to start your shell: %s", err)
	}
	defer guestProcess.Close()

	logger.Infof("Creating terminal...")
	terminal := terminal.New(pty, nil, conf)

	return terminal
}

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
	aminal = startAminal()
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
			log.Printf("key mask: %#b", mask)
			val := byte(key)
			log.Printf("key val: %#b", val)
			b := mask & val
			log.Printf("key byte: %#b", b)
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
			case 262:
				go func() { shellIn <- []byte("\u001b[C") }()
			case 263:
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
		log.Printf("Text input: %v\n", text)
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

	SetFont(ed.ActiveBuffer, 12)
	log.Println("Starting main loop")
	needsRedraw = true

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
			if needsRedraw {
				lasttime = glfw.GetTime()

				gfxMain(win, ctx, state)
				needsRedraw = false

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

//Aminal support functions

func getLogger(conf *config.Config) (*zap.SugaredLogger, error) {

	var logger *zap.Logger
	var err error
	if conf.DebugMode {
		logger, err = zap.NewDevelopment()
	} else {
		loggerConfig := zap.NewProductionConfig()
		loggerConfig.Encoding = "console"
		logger, err = loggerConfig.Build()
	}
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

func getConfig() *config.Config {
	showVersion := false
	ignoreConfig := false
	shell := ""
	debugMode := false
	slomo := false

	if flag.Parsed() == false {
		flag.BoolVar(&showVersion, "version", showVersion, "Output version information")
		flag.BoolVar(&ignoreConfig, "ignore-config", ignoreConfig, "Ignore user config files and use defaults")
		flag.StringVar(&shell, "shell", shell, "Specify the shell to use")
		flag.BoolVar(&debugMode, "debug", debugMode, "Enable debug logging")
		flag.BoolVar(&slomo, "slomo", slomo, "Render in slow motion (useful for debugging)")

		flag.Parse() // actual parsing and fetching flags from the command line
	}
	actuallyProvidedFlags := getActuallyProvidedFlags()

	var conf *config.Config
	if ignoreConfig {
		conf = &config.DefaultConfig
	} else {
		conf = loadConfigFile()
	}

	// Override values in the configuration file with the values specified in the command line, if any.
	if actuallyProvidedFlags["shell"] {
		conf.Shell = shell
	}

	if actuallyProvidedFlags["debug"] {
		conf.DebugMode = debugMode
	}

	if actuallyProvidedFlags["slomo"] {
		conf.Slomo = slomo
	}

	return conf
}

func loadConfigFile() *config.Config {

	usr, err := user.Current()
	if err != nil {
		fmt.Printf("Failed to get current user information: %s\n", err)
		return &config.DefaultConfig
	}

	home := usr.HomeDir
	if home == "" {
		return &config.DefaultConfig
	}

	places := []string{}

	places = append(places, filepath.Join(home, ".config/aminal/config.toml"))
	places = append(places, filepath.Join(home, ".aminal.toml"))

	for _, place := range places {
		if b, err := ioutil.ReadFile(place); err == nil {
			if c, err := config.Parse(b); err == nil {
				return c
			}

			fmt.Printf("Invalid config at %s: %s\n", place, err)
		}
	}

	if b, err := config.DefaultConfig.Encode(); err != nil {
		fmt.Printf("Failed to encode config file: %s\n", err)
	} else {
		err = os.MkdirAll(filepath.Dir(places[0]), 0744)
		if err != nil {
			fmt.Printf("Failed to create config file directory: %s\n", err)
		} else {
			if err = ioutil.WriteFile(places[0], b, 0644); err != nil {
				fmt.Printf("Failed to encode config file: %s\n", err)
			}
		}
	}

	return &config.DefaultConfig
}
