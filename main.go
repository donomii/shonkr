// Copyright 2014 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build darwin linux windows

// An app that draws a green triangle on a red background.
//
// Note: This demo is an early preview of Go 1.5. In order to build this
// program as an Android APK using the gomobile tool.
//
// See http://godoc.org/golang.org/x/mobile/cmd/gomobile to install gomobile.
//
// Get the basic example and use gomobile to build or install it on your device.
//
//   $ go get -d golang.org/x/mobile/example/basic
//   $ gomobile build golang.org/x/mobile/example/basic # will build an APK
//
//   # plug your Android device to your computer or start an Android emulator.
//   # if you have adb installed on your machine, use gomobile install to
//   # build and deploy the APK to an Android target.
//   $ gomobile install golang.org/x/mobile/example/basic
//
// Switch to your device or emulator to start the Basic application from
// the launcher.
// You can also run the application on your desktop by running the command
// below. (Note: It currently doesn't work on Windows.)
//   $ go install golang.org/x/mobile/example/basic && basic
package main

import "github.com/pkg/profile"
import "image/color"
import (
"golang.org/x/mobile/event/key"
    "io/ioutil"
    _ "strings"
    //"math"
    "net"
    "errors"
    "encoding/binary"
    "log"
    "runtime"

    "golang.org/x/mobile/app"
    "golang.org/x/mobile/event/lifecycle"
    "golang.org/x/mobile/event/paint"
    "golang.org/x/mobile/event/size"
    "golang.org/x/mobile/event/touch"
    "golang.org/x/mobile/exp/app/debug"
    "golang.org/x/mobile/exp/f32"
    "golang.org/x/mobile/exp/gl/glutil"
    "golang.org/x/mobile/gl"
    "fmt"
    "os"
    "time"
    "image"
    //"math/rand"
    _ "image/png"
    "github.com/donomii/sceneCamera"
)
import "github.com/go-gl/mathgl/mgl32"
        import "golang.org/x/mobile/exp/sensor"

var multiSample = uint(1)  //Make the internal pixel buffer larger to enable multisampling and eventually GL anti-aliasing
var clientWidth=uint(800*multiSample)
var clientHeight=uint(600*multiSample)
var u8Pix []uint8
var (
    startDrawing bool
    imageData image.Image
    imageBounds image.Rectangle
    images   *glutil.Images
    fps      *debug.FPS
    program  gl.Program
    position gl.Attrib
    u_Texture gl.Uniform
    a_TexCoordinate gl.Attrib
    colour gl.Attrib
    buf      gl.Buffer
    tbuf      gl.Buffer

    screenWidth int
    screenHeight int

    green  float32
    red  float32
    blue  float32
    touchX float32
    touchY float32
    selection int
    gallery []string
    reCalcNeeded bool
    prevTime int64
)

var scanOn = true
var vMeta map[string]vertexMeta
var triBuff[]byte
var vTrisf map[string][]float32
var vBuffs map[string]gl.Buffer

var vCols map[string][]byte
var vColsf map[string][]float32
var vColBuffs map[string]gl.Buffer

var trans  mgl32.Mat4
var theatreCamera  mgl32.Mat4
var transU gl.Uniform
var recursion int = 4
var threeD bool = false
var polyCount int
var clock float32 = 0.0
var Tex gl.Texture
var sceneCam *sceneCamera.SceneCamera

var viewAngle [3]float32


var texAlignData = f32.Bytes(binary.LittleEndian,
    0.0, 0.0, // top left
    0.0, 1.0, // top left
    1.0, 0.0, // top left
    0.0, 1.0, // top left
    1.0, 1.0, // top left
    1.0, 0.0, // top left
)


var triangleData = f32.Bytes(binary.LittleEndian,
    -1.0, 1.0, 0.0, // top lef
    -1.0, -1.0, 0.0, // bottom left
    1.0, 1.0, 0.0, // bottom right
    -1.0, -1.0, 0.0, // bottom right
    1.0, -1.0, 0.0, // top left
    1.0, 1.0, 0.0, // bottom right
//
    //0.0, 1.0, 0.0, // top left
    //0.0, 0.0, 0.0, // bottom left
    //0.0, 0.0, 0.2, // bottom right
    //0.0, 1.0, 0.2, // top right
    //0.0, 1.0, 0.0, // top left
    //0.0, 0.0, 0.2, // bottom right
)

var texData = f32.Bytes(binary.LittleEndian,
    0.0, 1.0, 1.0, // top left
    1.0, 0.0, 0.0, // bottom left
    1.0, 0.0, 0.0, // bottom right
    0.0, 0.0, 1.0, // bottom right
    1.0, 0.0, 0.0, // top left
    0.0, 0.0, 0.0, // bottom right
//
    0.0, 1.0, 0.0, // top left
    0.0, 0.0, 0.0, // bottom left
    0.0, 0.0, 1.0, // bottom right
    0.0, 1.0, 1.0, // top right
    0.0, 1.0, 0.0, // top left
    0.0, 0.0, 0.9, // bottom right
    0.0, 0.0, 0.9, // bottom right
    0.0, 1.0, 0.9, // top right
    0.0, 1.0, 0.0, // top left
    0.0, 0.0, 0.9, // bottom right
)


func do_profile() {
    //defer profile.Start(profile.MemProfile).Stop()
    //defer profile.Start(profile.TraceProfile).Stop()
    defer profile.Start(profile.CPUProfile).Stop()
    time.Sleep(60*time.Second)
}

func main() {
    gc.ActiveBuffer = NewBuffer()
    gc.ActiveBuffer.Formatter = NewFormatter()
    gc.ActiveBuffer.Data.Text=`
Welcome to the shonky editor
----------------------------

Shonkr started as an attempt to combine vi style editing with an accelerated OpenGL layout engine, and promptly went off the rails.

To edit a file, start shonkr from the command line:

    shonkr my_file.txt

shonkr has two modes, like Vi - a movement/editing mode, and an insert mode

Press Escape to leave insert mode.  You can then use (a small number) of the usual Vi keys

N Next Buffer
V Paste text from Clipboard
~ Save File

+ Increase font size
- Decrease font size

i - insert
a - insert after next letter
A - insert at EOL
$ - Skip to EOL
0   Start of line
^   Start of text on line

B Clear all caches
`;
    log.Printf("Starting main...")
    sceneCam = sceneCamera.New()
    runtime.GOMAXPROCS(2)
    app.Main(func(a app.App) {
        log.Printf("Starting app...")
        reCalcNeeded = true
        var glctx gl.Context
        var sz size.Event
        sensor.Notify(a)
        theatreCamera = mgl32.Ident4()
        trans = mgl32.Ident4()
        trans = trans.Mul4(mgl32.Translate3D(0.0, 0.0, 1.0))
        if threeD {
            trans = compose(trans, mgl32.Scale3D(1.6, 0.6,1.0))
        }
        theatreCamera = mgl32.LookAt(0.0, 0.0, 0.6, 0.0, 0.0, 0.0, 0.0, 1.0, 0.0)
        for e := range a.Events() {
            switch e := a.Filter(e).(type) {
            case sensor.Event:
                  delta := e.Timestamp - prevTime
                  prevTime = e.Timestamp
                  scale := float32(36000000.0/float32(delta))
                  sceneCam.ProcessEvent(e)


                  var sora_vec mgl32.Vec3   //The real sora
                  sora_vec = mgl32.Vec3{float32(e.Data[1])/scale, -float32(e.Data[0])/scale,float32(-e.Data[2])/scale/float32(3.14)}

                  if threeD {
                  } else {
                      theatreCamera = theatreCamera.Mul4(mgl32.Translate3D(sora_vec[1]/scale, -sora_vec[0]/scale, 0.0))
                  }
            case lifecycle.Event:
                switch e.Crosses(lifecycle.StageVisible) {
                case lifecycle.CrossOn:
                    glctx, _ = e.DrawContext.(gl.Context)
                    onStart(glctx)
                    sensor.Enable(sensor.Gyroscope, 10 * time.Millisecond)
                    a.Send(paint.Event{})
                case lifecycle.CrossOff:
                    sensor.Disable(sensor.Gyroscope)
                    onStop(glctx)
                    glctx = nil
                }
            case size.Event:
                sz = e
                reCalcNeeded = true
                screenWidth = sz.WidthPx*int(multiSample)
                clientWidth = uint(sz.WidthPx)*multiSample
                screenHeight = sz.HeightPx*int(multiSample)
                clientHeight = uint(sz.HeightPx)*multiSample
                reDimBuff(screenWidth,screenHeight)
                touchX = float32(sz.WidthPx /2)
                touchY = float32(sz.HeightPx * 9/10)
                if (sz.Orientation == size.OrientationLandscape) {
                    //threeD = true
                } else {
                    threeD = false
                }
            case paint.Event:
                if glctx == nil || e.External {
                    // As we are actively painting as fast as
                    // we can (usually 60 FPS), skip any paint
                    // events sent by the system.
                    continue
                }

                onPaint(glctx, sz)
                a.Publish()
                // Drive the animation by preparing to paint the next frame
                // after this one is shown.
                a.Send(paint.Event{})
            case key.Event:
                if e.Direction != key.DirRelease {
                    handleEvent(a, e)
                } else {
                    log.Println(e)
                //Well, this is interesting.  The escape key does not send a Press event, only a release event
                    if e.Code == key.CodeEscape {
                        gc.ActiveBuffer.InputMode = false
                        e.Code = key.CodeQ
                    }
                }
            case touch.Event:
                theatreCamera = mgl32.LookAt(0.0, 0.0, 0.1, 0.0, 0.0, -0.5, 0.0, 1.0, 0.0)
                if e.Type == touch.TypeBegin {
                    reCalcNeeded = true
                    selection++
                    if selection +1  > len(gallery) {
                        selection=0
                    }
                }
            }
        }
    })
}

var connectCh chan bool


func externalIP() (string, error) {
    ifaces, err := net.Interfaces()
    if err != nil {
        return "", err
    }
    for _, iface := range ifaces {
        if iface.Flags&net.FlagUp == 0 {
            continue // interface down
        }
        if iface.Flags&net.FlagLoopback != 0 {
            continue // loopback interface
        }
        addrs, err := iface.Addrs()
        if err != nil {
            return "", err
        }
        for _, addr := range addrs {
            var ip net.IP
            switch v := addr.(type) {
            case *net.IPNet:
                ip = v.IP
            case *net.IPAddr:
                ip = v.IP
            }
            if ip == nil || ip.IsLoopback() {
                continue
            }
            ip = ip.To4()
            if ip == nil {
                continue // not an ipv4 address
            }
            return ip.String(), nil
        }
    }
    return "", errors.New("are you connected to the network?")
}

func reDimBuff(x,y int) {
    dim := x*y*4
    u8Pix = make([]uint8, dim, dim)
}

var fname string

func NewFormatter() *FormatParams{
    return &FormatParams{&color.RGBA{1,1,1,255},0,0,0, 6.0,0,0, false}
}

func NewBuffer() *Buffer{
    buf := &Buffer{}
    buf.Data = &BufferData{}
    buf.Formatter = NewFormatter()
    buf.Data.Text = ""
    buf.Data.FileName = ""
    return buf
}

func onStart(glctx gl.Context) {
    gc.ActiveBufferId = 0
    if len(os.Args)>1 {
        fname = os.Args[1]
        log.Println("Loading file: ", fname)
        b, _ := ioutil.ReadFile(fname)
        gc.ActiveBuffer = NewBuffer()
        gc.ActiveBuffer.Data.Text = string(b)
        gc.ActiveBuffer.Data.FileName = fname
        gc.BufferList = append(gc.BufferList, gc.ActiveBuffer)
    }
     if len(os.Args)>2 {
        fname = os.Args[2]
        log.Println("Loading file: ", fname)
        b, _ := ioutil.ReadFile(fname)
        buf := NewBuffer()
        buf.Data.Text = string(b)
        buf.Data.FileName = fname
        buf.Formatter.TailBuffer = true
        gc.BufferList = append(gc.BufferList, buf)
    }
    for i:=0;i<5;i++ {
        buf := NewBuffer()
        gc.BufferList = append(gc.BufferList, buf)
    }
    gc.BufferList[2].Data = gc.BufferList[1].Data
    log.Printf("Onstart callback...")
    reDimBuff(int(clientWidth),int(clientHeight))
    var err error
    program, err = glutil.CreateProgram(glctx, vertexShader, fragmentShader)
    if err != nil {
        log.Printf("error creating GL program: %v", err)
        os.Exit(1)
        return
    }


    position = glctx.GetAttribLocation(program, "position")
    a_TexCoordinate = glctx.GetAttribLocation(program, "a_TexCoordinate")
    transU = glctx.GetUniformLocation(program, "transform")
    u_Texture = glctx.GetUniformLocation(program, "u_Texture")
    //fmt.Println("Creating buffers")

    buf = glctx.CreateBuffer()
    glctx.BindBuffer(gl.ARRAY_BUFFER, buf)
    //fmt.Printf("triangleData: %V\n", triangleData)
    glctx.BufferData(gl.ARRAY_BUFFER, triangleData, gl.STATIC_DRAW)

    tbuf = glctx.CreateBuffer()
    glctx.BindBuffer(gl.ARRAY_BUFFER, tbuf)
    //fmt.Printf("texAlignData: %V\n", texAlignData)
    glctx.BufferData(gl.ARRAY_BUFFER, texAlignData, gl.STATIC_DRAW)


    Tex = glctx.CreateTexture()
    glctx.BindTexture(gl.TEXTURE_2D, Tex)

    glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
    glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);


}

func onStop(glctx gl.Context) {
    log.Printf("Stopping...")
    os.Exit(0)
    //glctx.DeleteProgram(program)
    //glctx.DeleteBuffer(buf)
    //fps.Release()
    //images.Release()
}

func transpose( m mgl32.Mat4) mgl32.Mat4{
    var r mgl32.Mat4
    for i, v := range []int{0,4,8,12,1,5,9,13,2,6,10,14,3,7,11,15} {
        r[i] = m[v]
    }
    //fmt.Println(r)
    return r
}

func onPaint(glctx gl.Context, sz size.Event) {
    for i, _:= range u8Pix {
        u8Pix[i] = 0
    }
    RenderPara(gc.ActiveBuffer.Formatter, 5, 5, screenWidth, screenHeight, u8Pix, gc.ActiveBuffer.Data.Text, true, true, true)
    glctx.Enable(gl.BLEND)
    glctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
    glctx.Enable( gl.DEPTH_TEST );
    glctx.DepthFunc( gl.LEQUAL );
    glctx.DepthMask(true)
    glctx.Clear(gl.COLOR_BUFFER_BIT|gl.DEPTH_BUFFER_BIT)
    glctx.ClearColor(255,255,255,255)
    glctx.UseProgram(program)


    //outBytes := make([]byte, len(u8Pix))
    //for i:=2 ; i<len(u8Pix)-2; i++ {
        ////n := (u8Pix[i-2] + u8Pix[i-1] + u8Pix[i] + u8Pix[i+1] + u8Pix[i+2])/5
        //n := (u8Pix[i-1] + u8Pix[i] + u8Pix[i+1])/3
        //m := int(math.Ceil(math.Mod(float64(i), 4)))
        //log2Buff(fmt.Sprintf("n: %v, m: %v, orig: %v\n", n,m, u8Pix[i]))
        ////fmt.Printf("n: %v, m: %v, orig: %v\n", n,m, u8Pix[i])
        //if n> u8Pix[i] {
            //if  m != 3 {
                //outBytes[i] = n
            //} else {
                //outBytes[i] = 255
            //}
        //} else {
            //outBytes[i] = u8Pix[i]
        //}
    //}

    glctx.TexImage2D(gl.TEXTURE_2D, 0, int(clientWidth), int(clientHeight), gl.RGBA, gl.UNSIGNED_BYTE, u8Pix)


    var view mgl32.Mat4
    if threeD {
        view = compose3(mgl32.Perspective(55, float32(screenWidth)/float32(screenHeight), 0.1, 2048.0), sceneCam.ViewMatrix(), trans)
    } else {
        view = compose(theatreCamera, trans)
    }
    glctx.UniformMatrix4fv(transU, view[0:16])

    glctx.BindBuffer(gl.ARRAY_BUFFER, buf)
    glctx.EnableVertexAttribArray(position)
    glctx.VertexAttribPointer(position, 3, gl.FLOAT, false, 0, 0)


    glctx.BindBuffer(gl.ARRAY_BUFFER, tbuf)
    glctx.EnableVertexAttribArray(a_TexCoordinate)
    glctx.VertexAttribPointer(a_TexCoordinate, 2, gl.FLOAT, false, 0, 0)

    glctx.ActiveTexture(gl.TEXTURE0);
    // Bind the texture to this unit.
    glctx.BindTexture(gl.TEXTURE_2D, Tex);
    // Tell the texture uniform sampler to use this texture in the shader by binding to texture unit 0.
    glctx.Uniform1i(u_Texture, 0);

    glctx.Viewport(0,0, sz.WidthPx/2, sz.HeightPx)
    glctx.DrawArrays(gl.TRIANGLES, 0, 6)


    for i, _:= range u8Pix {
        u8Pix[i] = 0
    }
    RenderPara(gc.BufferList[1].Formatter, 5, 5, screenWidth, screenHeight, u8Pix, gc.BufferList[1].Data.Text, true, true, false)

    //for i:=1 ; i<len(u8Pix)-1; i++ {
        //outBytes[i] = u8Pix[i] // (u8Pix[i-1] + 2*u8Pix[i] +u8Pix[i+1])/4
    //}


    glctx.TexImage2D(gl.TEXTURE_2D, 0, int(clientWidth), int(clientHeight), gl.RGBA, gl.UNSIGNED_BYTE, u8Pix)
    glctx.Viewport(sz.WidthPx/2,0, sz.WidthPx/2, sz.HeightPx)
    glctx.DrawArrays(gl.TRIANGLES, 0, 6)
    glctx.DisableVertexAttribArray(position)

}
type GlobalConfig struct {
    ActiveBuffer   *Buffer
    ActiveBufferId int
    BufferList     []*Buffer
}

type BufferData struct {
    Text string     //FIXME rename Buffer to View, have proper text buffer manager
    FileName string
}

type Buffer struct {
    Data    *BufferData
    InputMode bool
    Formatter *FormatParams
}

var gc GlobalConfig


type vertexMeta struct {
    coordsPerVertex int
    vertexCount     int
}

const (
    coordsPerVertex = 3
    vertexCount     = 3
)


const vertexShader = `#version 100
uniform mat4 transform;

attribute vec2 a_TexCoordinate; // Per-vertex texture coordinate information we will pass in.
varying vec2 v_TexCoordinate;   // This will be passed into the fragment shader.

    attribute vec4 position;
    varying vec4 color;
void main() {
        gl_Position = transform * position;
        color = vec4(1.0,1.0,1.0,1.0);
        // Pass through the texture coordinate.
        v_TexCoordinate = a_TexCoordinate;
}
`

const fragmentShader = `#version 100
precision mediump float;
varying vec4 color;
uniform sampler2D u_Texture;    // The input texture.
varying vec2 v_TexCoordinate; // Interpolated texture coordinate per fragment.
void main() {
    //gl_FragColor = color;
    gl_FragColor = texture2D(u_Texture, v_TexCoordinate);
}`



func compose (a, b mgl32.Mat4) mgl32.Mat4 {
return a.Mul4(b)
}

func compose3 (a, b, c mgl32.Mat4) mgl32.Mat4 {
    t := b.Mul4(c)
return a.Mul4(t)
}

func checkGlErr(glctx gl.Context) {
    err := glctx.GetError()
    if (err>0) {
        fmt.Printf("GLerror: %v\n", err)
        panic("GLERROR")
    }
}
