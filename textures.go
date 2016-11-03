// textures
package main

import "math"
import (
	"strings"
	"unicode"
	"unicode/utf8"
	"github.com/kardianos/osext"
	//sysFont "golang.org/x/mobile/exp/font"
	"io/ioutil"
	//"bytes"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"

	"github.com/go-gl/mathgl/mgl32"

	"golang.org/x/mobile/gl"

	"image"
	"image/color"
	"image/png"
)

type Thunk func()

var (
	rtt_frameBuff gl.Framebuffer
	rtt_tex       gl.Texture
)

func screenShot(glctx gl.Context, filename string) {
	log.Printf("Saving width: %v, height: %v\n", screenWidth, screenHeight)
	saveBuff(uint(screenWidth), uint(screenHeight), copyScreen(glctx, int(screenWidth), int(screenHeight)), filename)
}

//Copies an image to a correctly-packed texture data array.
//
//Returns the array, modified in place.  If u8Pix is nil or texWidth is 0, it creates a new texture array and returns that.  Texture is assumed to be square.
func paintTexture(img image.Image, u8Pix []uint8, clientWidth uint) []uint8 {
	bounds := img.Bounds()
	newW := bounds.Max.X
	newH := bounds.Max.Y

	//if uint(newW) != clientWidth || uint(newH) != clientWidth {
	if (uint(newW) > clientWidth) || (uint(newH) > clientWidth) {
		panic(fmt.Sprintf("ClientWidth (%v) does not match image width(%v) and height(%v)", clientWidth, newW, newH))
	}
	if u8Pix == nil {
		dim := clientWidth*clientWidth*4 + 4
		u8Pix = make([]uint8, dim, dim)
	}

	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			// A color's RGBA method returns values in the range [0, 65535].
			start := uint(y)*clientWidth*4 + uint(x)*4
			u8Pix[start] = uint8(r * 255 / 65535)
			u8Pix[start+1] = uint8(g * 255 / 65535)
			u8Pix[start+2] = uint8(b * 255 / 65535)
			u8Pix[start+3] = uint8(a * 255 / 65535)
		}
	}
	return u8Pix
}

func copyScreen(glctx gl.Context, clientWidth, clientHeight int) []byte {
	buff := make([]byte, clientWidth*clientHeight*4, clientWidth*clientHeight*4)
	//fmt.Printf("reading pixels: ")
	//glctx.BindFramebuffer(gl.FRAMEBUFFER, rtt_frameBuff)
	glctx.ReadPixels(buff, 0, 0, clientWidth, clientHeight, gl.RGBA, gl.UNSIGNED_BYTE)
	checkGlErr(glctx)
	glctx.BindFramebuffer(gl.FRAMEBUFFER, gl.Framebuffer{0})
	return buff
}

func copyFrameBuff(glctx gl.Context, clientWidth, clientHeight int) []byte {
	buff := make([]byte, clientWidth*clientHeight*4, clientWidth*clientHeight*4)
	//fmt.Printf("reading pixels: ")
	glctx.BindFramebuffer(gl.FRAMEBUFFER, rtt_frameBuff)
	glctx.ReadPixels(buff, 0, 0, clientWidth, clientHeight, gl.RGBA, gl.UNSIGNED_BYTE)
	checkGlErr(glctx)
	glctx.BindFramebuffer(gl.FRAMEBUFFER, gl.Framebuffer{0})
	return buff
}

func saveImage(m *image.RGBA, filename string) {
	f, _ := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	defer f.Close()
	png.Encode(f, m)
}

func saveBuff(texWidth, texHeight uint, buff []byte, filename string) {
	f, _ := os.OpenFile(filename, os.O_TRUNC|os.O_CREATE|os.O_WRONLY, 0666)
	defer f.Close()
	m := image.NewNRGBA(image.Rectangle{image.Point{0, 0}, image.Point{int(texWidth), int(texHeight)}})
	if buff != nil {
		//fmt.Printf("readpixels: %V", buff)
		for y := uint(0); y < texWidth; y++ {
			for x := uint(0); x < texHeight; x++ {
				i := (x + y*texWidth) * 4
				m.Set(int(x), int(texHeight-y), color.NRGBA{uint8(buff[i]), uint8(buff[i+1]), uint8(buff[i+2]), 255})
				//if buff[i]>0 { fmt.Printf("Found colour\n") }
				//if buff[i+1]>0 { fmt.Printf("Found colour\n") }
				//if buff[i+2]>0 { fmt.Printf("Found colour\n") }
			}
		}
	}
	png.Encode(f, m)
}

func Rtt(glctx gl.Context, texWidth, texHeight int, thunk Thunk) {
	glctx.BindFramebuffer(gl.FRAMEBUFFER, rtt_frameBuff)
	glctx.Viewport(0, 0, texWidth, texHeight)
	glctx.ActiveTexture(gl.TEXTURE0)
	glctx.BindTexture(gl.TEXTURE_2D, rtt_tex)
	//draw here the content you want in the texture
	log.Printf("Framebuffer status: %v\n", glctx.CheckFramebufferStatus(gl.FRAMEBUFFER))
	trans = mgl32.Ident4()

	//rtt_tex is now a texture with the drawn content

	glctx.Enable(gl.BLEND)
	glctx.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	glctx.Enable(gl.DEPTH_TEST)
	glctx.DepthFunc(gl.LEQUAL)
	glctx.DepthMask(true)
	glctx.ClearColor(0, 0, 0, 1)
	glctx.UseProgram(program) //FIXME - may cause graphics glitches
	polyCount = 0
	clock += 0.001
	glctx.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	thunk()

	glctx.Flush()

	buff := copyFrameBuff(glctx, texWidth, texHeight)
	saveBuff(uint(texWidth), uint(texHeight), buff, "x.png")
	glctx.BindTexture(gl.TEXTURE_2D, gl.Texture{0})
	glctx.BindFramebuffer(gl.FRAMEBUFFER, gl.Framebuffer{0})
	log.Println("Finished Rtt")
	fmt.Printf("done \n")
}

func dumpBuff(buff []uint8, width, height uint) {
	log.Printf("Dumping buffer with width, height %v,%v\n", width, height)
	for y := uint(0); y < height; y++ {
		for x := uint(0); x < width; x++ {
			i := (x + y*width) * 4
			//log.Printf("Index: %v\n", i)
			if buff[i] > 128 {
				fmt.Printf("I")
			} else {
				fmt.Printf("_")
			}
		}
		fmt.Println("")
	}
}

func string2Tex(glctx gl.Context, str string, tSize float64, glTex gl.Texture) {

	img, _ := DrawStringRGBA(tSize, color.RGBA{255, 255, 255, 255}, str)
	saveImage(img, "texttest.png")
	w := 128 //img.Bounds().Max.X  //FIXME
	buff := paintTexture(img, nil, uint(w))
	dumpBuff(buff, uint(w), uint(w))
	uploadTex(glctx, glTex, w, w, buff)
}

func uploadTex(glctx gl.Context, glTex gl.Texture, w, h int, buff []uint8) {
	glctx.BindTexture(gl.TEXTURE_2D, glTex)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	glctx.TexImage2D(gl.TEXTURE_2D, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, buff)
	glctx.GenerateMipmap(gl.TEXTURE_2D)
}

//Creates a new framebuffer and texture, with the texture attached to the frame buffer
func glGenTextureFromFramebuffer(glctx gl.Context, w, h int) (gl.Framebuffer, gl.Texture) {
	f := glctx.CreateFramebuffer()
	glctx.BindFramebuffer(gl.FRAMEBUFFER, f)
	glctx.ActiveTexture(gl.TEXTURE0)
	t := glctx.CreateTexture()
	log.Printf("Texture created: %v", t)

	glctx.BindTexture(gl.TEXTURE_2D, t)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	glctx.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)

	glctx.TexImage2D(gl.TEXTURE_2D, 0, w, h, gl.RGBA, gl.UNSIGNED_BYTE, nil)

	glctx.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, t, 0)

	/*
	   depthbuffer := glctx.CreateRenderbuffer()
	   glctx.BindRenderbuffer(gl.RENDERBUFFER, depthbuffer)
	   glctx.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT16, w, h)
	   glctx.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, depthbuffer)
	*/

	status := glctx.CheckFramebufferStatus(gl.FRAMEBUFFER)
	if status != gl.FRAMEBUFFER_COMPLETE {
		fmt.Printf("Framebuffer status: %v\n", status)
		os.Exit(1)
	}
	glctx.BindFramebuffer(gl.FRAMEBUFFER, gl.Framebuffer{0})
	return f, t
}

var renderCache map[string]*image.RGBA
var faceCache map[string]*font.Face
var fontCache map[string]*truetype.Font

func clearAllCaches() {
	renderCache = map[string]*image.RGBA{}
	faceCache = map[string]*font.Face{}
	fontCache = map[string]*truetype.Font{}
}
func DrawStringRGBA(txtSize float64, fontColor color.RGBA, txt string) (*image.RGBA, *font.Face) {
    cacheKey := fmt.Sprintf("%v,%v,%v", txtSize, fontColor, txt)
    if renderCache == nil {
        renderCache = map[string]*image.RGBA{}
    }
    if faceCache == nil {
        faceCache = map[string]*font.Face{}
    }
    im, ok := renderCache[cacheKey]
    face, ok1 := faceCache[cacheKey]
    if ok && ok1 {
        return im, face
    }
    txtFont := LoadGameFont("f1.ttf")
    d := &font.Drawer{
        Src: image.NewUniform(fontColor), // 字体颜色
        Face: truetype.NewFace(txtFont, &truetype.Options{
            Size:    txtSize,
            DPI:     72,
            Hinting: font.HintingNone,
        }),
    }
    fface := d.Face
    glyph, _ := utf8.DecodeRuneInString(txt)
    fuckedRect, _, _ := fface.GlyphBounds(glyph)
    // letterWidth := fixed2int(fuckedRect.Max.X)
    Xadj := fixed2int(fuckedRect.Min.X)
    if Xadj<0 { Xadj = Xadj * -1 }
    // fuckedRect, _, _ = fface.GlyphBounds(glyph)
    // letterHeight := fixed2int(fuckedRect.Max.Y)
    //]
    rect := image.Rect(0, 0, d.MeasureString(txt).Ceil()*2, int(txtSize)*4)
    //rect := image.Rect(0, 0, 30, 30)
    rgba := image.NewRGBA(rect)
    d.Dst = rgba

    d.Dot = fixed.Point26_6{
        X: fixed.I(Xadj),
        Y: fixed.I(rect.Max.Y*2/3), //fixed.I(rect.Max.Y/3), //rect.Max.Y*2/3),
    }
    d.DrawString(txt)
    renderCache[cacheKey] = rgba
    faceCache[cacheKey] = &d.Face
    //imgBytes := rgba.Pix
    //for i,v := range imgBytes {
       //imgBytes[i] = 255 - v
    //}
    return rgba, &d.Face
}

func LoadGameFont(fileName string) *truetype.Font {

    if fontCache == nil {
        fontCache = map[string]*truetype.Font{}
    }
    im, ok := fontCache[fileName]
    if ok {
        return im
    }

	//fontBytes := sysFont.Monospace()
	//fontBytes := sysFont.Default()
	//f := bytes.NewReader(fontBytes)

	var f io.Reader
	if len(fileName) <= 0 {
		folderPath, err := osext.ExecutableFolder()
		if err != nil {
			log.Fatal(err)
		}
		file, err := os.Open(fmt.Sprintf("%v%v%v", folderPath, string(os.PathSeparator), "f1.txt"))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		f = file
	} else {
		//log.Println(fileName)
		folderPath, err := osext.ExecutableFolder()
		if err != nil {
			log.Fatal(err)
		}
		file, err := os.Open(fmt.Sprintf("%v%v%v", folderPath, string(os.PathSeparator), fileName))
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()
		f = file

	}
	fontBytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		panic(err)
	}

	txtFont, err1 := truetype.Parse(fontBytes)
	if err1 != nil {
		log.Println(err1)
		panic(err1)
	}

    fontCache[fileName] = txtFont
	return txtFont
}

func dumpBuffer(b *Buffer) {
	log2Buff(fmt.Sprintf(`
FileName: %v,
Active Buffer: %v,
StartChar: %v,
LastChar: %v,
Cursor: %v,
Tail: %v,
Font Size: %v,
Screen Width: %v,
Screen Height: %v
`, b.Data.FileName, gc.ActiveBufferId, b.Formatter.FirstDrawnCharPos, b.Formatter.LastDrawnCharPos, b.Formatter.Cursor, b.Formatter.TailBuffer, b.Formatter.FontSize, screenWidth, screenHeight))
}

type FormatParams struct {
	Colour            *color.RGBA
	Line              int
	Cursor            int
	StartLinePos      int //Updated during render, holds the closest start of line, including soft line breaks
	FontSize          float64
	FirstDrawnCharPos int //The first character to draw on the screen.  Anything before this is ignored
	LastDrawnCharPos  int //The last character that we were able to fit on the screen
	TailBuffer        bool
}

func drawCursor(xpos, ypos, height int, u8Pix []byte) {
	colour := byte(0)
	for xx := int(0); xx < 3; xx++ {
		for yy := int(0); yy < height; yy++ {
			offset := uint(yy+ypos)*clientWidth*4 + uint(xx+xpos)*4
			//log.Printf("Drawpos: %v", offset)
			if offset >= 0 && offset < uint(len(u8Pix)) {
				u8Pix[offset] = colour
				u8Pix[offset+1] = colour
				u8Pix[offset+2] = colour
				u8Pix[offset+3] = 255
			}
		}
	}
}

func searchBackPage(txtBuf string, orig_f *FormatParams) int {
	input := *orig_f
	x := input.StartLinePos
	newLastDrawn := input.LastDrawnCharPos
	for x = input.Cursor; x > 0 && input.FirstDrawnCharPos < newLastDrawn; x = scanToPrevLine(txtBuf, x) {
		f := input
		f.FirstDrawnCharPos = x
		RenderPara(&f, 0, 0, screenWidth, screenHeight, nil, txtBuf, false, false, false)
		newLastDrawn = f.LastDrawnCharPos
	}
	return x
}

//Check and correct formatparams to make sure e.g. cursor is always on the screen
func sanityCheck(f *FormatParams, txt string) {
	if f.Cursor < 0 {
		f.Cursor = 0
	}
	if f.Cursor >= len(txt)-1 {
		f.Cursor = len(txt) - 1
	}
	if f.FirstDrawnCharPos < 0 {
		f.FirstDrawnCharPos = 0
	}
	if f.FirstDrawnCharPos >= len(txt)-1 {
		f.FirstDrawnCharPos = len(txt) - 1
	}

}

func fixed2int(n fixed.Int26_6) int {
	return n.Round()
}

func log2Buff(s string) {
	gc.StatusBuffer.Data.Text = s
}


func RenderPara(f *FormatParams, orig_xpos, orig_ypos, maxX, maxY int, u8Pix []uint8, text string, transparent bool, doDraw bool, showCursor bool) {
	if f.TailBuffer {
		//f.Cursor = len(text)
		//scrollToCursor(f, text)  //Use pageup function, once it is fast enough
	}
	//log.Printf("Cursor: %v\n", f.Cursor)
	letters := strings.Split(text, "")
	letters = append(letters, " ")
	xpos := orig_xpos
	ypos := orig_ypos
	//xAdv := 1
	//yAdv := 0
	//lfX :=0
	//lfY := 1
	orig_fontSize := f.FontSize
	defer func() {
		f.FontSize = orig_fontSize
		sanityCheck(f, text)
	}()
	maxHeight := 0
	wobblyMode := false
	if f.Cursor > len(letters) {
		f.Cursor = len(letters)
	}
	//sanityCheck(f,txt)
	for i, v := range letters {
		if i < f.FirstDrawnCharPos {
			continue
		}
		if (f.Cursor == i) && doDraw {
			drawCursor(xpos, ypos, maxHeight, u8Pix)
		}
		if i >= len(letters)-1 {
			continue
		}
		if unicode.IsUpper([]rune(v)[0]) {
			//if i>0 && letters[i-1] == " " {
			//f.Colour = &color.RGBA{255,0,0,255}
			//f.FontSize = f.FontSize*1.2
			////log.Printf("Oversize start for %v at %v\n", v, i)
			//} else {
			//f.Colour = &color.RGBA{1,1,1,255}
			//}
			f.Colour = &color.RGBA{255, 1, 1, 255}
		} else {
			f.Colour = &color.RGBA{1, 32, 1, 255}
		}
		if (string(text[i]) == " ") || (string(text[i]) == "\n") {
			f.FontSize = orig_fontSize
			//log.Printf("Oversize end for %v at %v\n", v, i)
		}
		if string(text[i]) == "\n" {
			//ypos = ypos + maxHeight
			xpos = xpos + maxHeight
			//maxHeight=0
			//xpos = orig_xpos FIXME
            ypos = orig_ypos
			f.Line++
			f.StartLinePos = i
			if f.Cursor == i && showCursor {
				drawCursor(xpos, ypos, maxHeight, u8Pix)
			}
		} else {
			//if f.Cursor <= f.Line {
			//if f.Cursor >= f.FirstDrawnCharPos {
			if i >= f.FirstDrawnCharPos {
				ytweak := 0
				if wobblyMode {
					ytweak = int(math.Sin(float64(xpos)) * 5.0)
				}
				img, face := DrawStringRGBA(f.FontSize, *f.Colour, v)
				XmaX, YmaX := img.Bounds().Max.X, img.Bounds().Max.Y
				imgBytes := Rotate270(XmaX, YmaX, img.Pix)
				XmaX, YmaX = YmaX, XmaX
				//imgBytes := img.Pix
				fa := *face
				//glyph, _ := utf8.DecodeRuneInString(v)
				//letterWidth_F, _ := fa.GlyphAdvance(glyph)
				//letterWidth := fixed2int(letterWidth_F)
				//fuckedRect, _, _ := fa.GlyphBounds(glyph)
				//letterHeight := fixed2int(fuckedRect.Max.Y)
				letterHeight := fixed2int(fa.Metrics().Height)
				//letterWidth = XmaX
				//letterHeight = letterHeight

				if xpos+XmaX > maxX {
					/* 
                    ypos = ypos + maxHeight
					maxHeight = 0
					xpos = orig_xpos
					f.Line++
					f.StartLinePos = i
                    */
					f.LastDrawnCharPos = i - 1
					return
				}

				if ypos+YmaX+ytweak+1 > maxY {
                    /*
					f.LastDrawnCharPos = i - 1
					return
                    */
                    xpos = xpos + maxHeight
					maxHeight = 0
					ypos = orig_ypos
					f.Line++
					f.StartLinePos = i
				}

				if doDraw {
					//PasteImg(img, xpos, ypos + ytweak, u8Pix, transparent)
					PasteBytes(XmaX, YmaX, imgBytes, xpos, ypos+ytweak, int(clientWidth), int(clientHeight), u8Pix, transparent)
				}

				if f.Cursor == i && showCursor {
					drawCursor(xpos, ypos, maxHeight, u8Pix)
				}

				f.LastDrawnCharPos = i
				maxHeight = MaxI(maxHeight, letterHeight)
				//xpos = xpos+ po2/2
				newXpos := xpos
				newYpos := ypos
				//for ii:=0; ii<letterWidth; ii++ {
				//newXpos =  newXpos + xAdv
				//newYpos =  newYpos + yAdv
				//}

				xpos = newXpos
				ypos = newYpos

				//xpos += letterWidth
				ypos += maxHeight
				//log2Buff(fmt.Sprintf("Xadv: %v, Yadv: %v\n", XmaX, YmaX ))
				//log.Printf("Xadv: %v, Yadv: %v\n", XmaX, YmaX )
			}
		}
	}
	sanityCheck(f, text)
}

func MaxI(a, b int) int {
	if a > b {
		return a
	}
	return b
}

//PasteBytes
//
// Takes a bag of bytes, and some dimensions, and pastes it into another bag of bytes
func PasteBytes(srcWidth, srcHeight int, srcBytes []byte, xpos, ypos, dstWidth, dstHeight int, u8Pix []uint8, transparent bool) {
	//log.Printf("Copying source image (%v,%v) into destination image (%v,%v) at point (%v, %v)\n", srcWidth, srcHeight, dstWidth, dstHeight, xpos, ypos)
	startDrawing = true
	bpp := 4 //bytes per pixel

	for i := 0; i < srcHeight; i++ {
		if transparent {
			for j := 0; j < srcWidth; j++ {
				srcOff := i*srcWidth*4 + j*4
				dstOff := (ypos+i)*dstWidth*bpp + xpos*bpp + j*bpp
				// if(i==j) {
				//     u8Pix[dstOff] = 255
				//     u8Pix[dstOff+3] = 255
				// }
				r := srcBytes[srcOff+0]
				g := srcBytes[srcOff+1]
				b := srcBytes[srcOff+2]

				dstR := u8Pix[srcOff+0]
				dstG := u8Pix[srcOff+1]
				dstB := u8Pix[srcOff+2]

				srcA := srcBytes[srcOff+3]
				dstA := u8Pix[dstOff+3]

				outA := srcA + dstA*(1-srcA)

				outR := byte(0)
				outG := byte(0)
				outB := byte(0)
				if outA > 0 {
					outR = (r*srcA + dstR*dstA*(1-srcA)) / outA
					outG = (g*srcA + dstG*dstA*(1-srcA)) / outA
					outB = (b*srcA + dstB*dstA*(1-srcA)) / outA
				} else {
					outR = 0
					outG = 0
					outB = 0
				}
				if srcBytes[i*srcWidth*4+j*4] > u8Pix[(ypos+i)*dstWidth*bpp+xpos*bpp+j*bpp] {
					//log2Buff(fmt.Sprintf("Source: (%v,%v), destination: (%v,%v)\n", j,i,xpos+j, ypos+i))
					copy(u8Pix[dstOff:dstOff+4], srcBytes[srcOff:srcOff+4])
				}
				u8Pix[srcOff+0] = outR
				u8Pix[srcOff+1] = outG
				u8Pix[srcOff+2] = outB
			}
		} else {
			srcOff := i * srcWidth * 4
			dstOff := (ypos+i)*dstWidth*bpp + xpos*bpp
			copy(u8Pix[dstOff:dstOff+4*srcWidth], srcBytes[srcOff:srcOff+4*srcWidth]) //FIXME move this outside the line loop so we can copy entire lines in one call
		}
	}
}

func PasteImg(img *image.RGBA, xpos, ypos int, u8Pix []uint8, transparent bool) {
	po2 := uint(MaxI(NextPo2(img.Bounds().Max.X), NextPo2(img.Bounds().Max.Y)))
	//log.Printf("Chose texture size: %v\n", po2)
	wordBuff := paintTexture(img, nil, po2)
	bpp := uint(4) //bytes per pixel

	h := img.Bounds().Max.Y
	w := uint(img.Bounds().Max.X)
	for i := uint(0); i < uint(h); i++ {
		for j := uint(0); j < w; j++ {
			if (wordBuff[i*po2*4+j*4] > 128) || !transparent {
				u8Pix[(uint(ypos)+i)*clientWidth*bpp+uint(xpos)*bpp+j*bpp] = wordBuff[i*po2*4+j*4]
				u8Pix[(uint(ypos)+i)*clientWidth*bpp+uint(xpos)*bpp+j*bpp+1] = wordBuff[i*po2*4+j*4+1]
				u8Pix[(uint(ypos)+i)*clientWidth*bpp+uint(xpos)*bpp+j*bpp+2] = wordBuff[i*po2*4+j*4+2]
				u8Pix[(uint(ypos)+i)*clientWidth*bpp+uint(xpos)*bpp+j*bpp+3] = wordBuff[i*po2*4+j*4+3]
			}
		}
	}
}

func PasteText(tSize float64, xpos, ypos int, text string, u8Pix []uint8, transparent bool) {
	img, _ := DrawStringRGBA(tSize, color.RGBA{255, 255, 255, 255}, text)
	po2 := uint(MaxI(NextPo2(img.Bounds().Max.X), NextPo2(img.Bounds().Max.Y)))
	//log.Printf("Chose texture size: %v\n", po2)
	wordBuff := paintTexture(img, nil, po2)
	startDrawing = true
	bpp := uint(4) //bytes per pixel

	h := img.Bounds().Max.Y
	w := uint(img.Bounds().Max.X)
	for i := uint(0); i < uint(h); i++ {
		for j := uint(0); j < w; j++ {
			if (wordBuff[i*po2*4+j*4] > 128) || !transparent {
				u8Pix[(uint(ypos)+i)*clientWidth*bpp+uint(xpos)*bpp+j*bpp] = wordBuff[i*po2*4+j*4]
				u8Pix[(uint(ypos)+i)*clientWidth*bpp+uint(xpos)*bpp+j*bpp+1] = wordBuff[i*po2*4+j*4+1]
				u8Pix[(uint(ypos)+i)*clientWidth*bpp+uint(xpos)*bpp+j*bpp+2] = wordBuff[i*po2*4+j*4+2]
				u8Pix[(uint(ypos)+i)*clientWidth*bpp+uint(xpos)*bpp+j*bpp+3] = wordBuff[i*po2*4+j*4+3]
			}
		}
	}
}

func NextPo2(n int) int {
	return int(math.Pow(2, math.Ceil(math.Log2(float64(n)))))
}

func Rotate270(srcW, srcH int, src []byte) []byte {
	//log.Printf("Rotating image (%v,%v)\n",srcW, srcH)
	dstW := srcH
	dstH := srcW
	dst := make([]byte, dstW*dstH*4)

	for dstY := 0; dstY < dstH; dstY++ {
		for dstX := 0; dstX < dstW; dstX++ {
			srcX := dstY
			srcY := dstW - dstX - 1
			//srcY := dstX

			srcOff := srcY*srcW*4 + srcX*4
			dstOff := dstY*dstW*4 + dstX*4

			copy(dst[dstOff:dstOff+4], src[srcOff:srcOff+4])
		}
	}

	return dst
}
