package main
import (
"math"
"unicode/utf8"
"github.com/donomii/nucular/rect"
 //"image"
 "github.com/donomii/nucular"
 "image/color"
 nstyle "github.com/donomii/nucular/style"
 "github.com/donomii/glim"
// "github.com/disintegration/imaging"
// "image/draw"
 "strings"
 "unicode"
)

func DrawString(win *nucular.Window,x1,y1,x2,y2, w, h int, buff []byte) {
	img := glim.ImageToGFormatRGBA(w,h,buff)
win.Cmds().DrawImage(rect.Rect{50, 100, 200, 200}, img)
}


func updatefn(w *nucular.Window) {
	w.Row(30).Dynamic(1)
	w.Label("Dynamic fixed column layout with generated position and size (LayoutRowDynamic):", "LC")
	w.Row(30).Dynamic(1)
	w.LabelColored("Hello", "LC", color.RGBA{255,255,255,255})
	img, _ := glim.DrawStringRGBA(9.6, color.RGBA{255,255,255,255}, "Hello again", "f1.ttf")
	newH := img.Bounds().Max.Y
	w.Row(newH).Dynamic(1)
	w.Image(img)
	img2, W, H := glim.GFormatToImage(img, nil, 0, 0)
	img2 = glim.MakeTransparent(img2, color.RGBA{0,0,0,0})
	img3 := glim.Rotate270(W, H, img2)
	img4 := glim.ImageToGFormatRGBA(H, W, img3)
	img5 := img4
	w.Image(img5)
	w.Cmds().DrawImage(rect.Rect{50, 100, 200, 200}, img5)
	f := glim.NewFormatter()
	f.FontSize = 9.6
	nw := 800
	nh := 500
	buff := make([]byte, nw*nh*4)
	
    glim.RenderPara(f , 10, 15, 0, 0, nw, nh, nw, nh, 1, 1, buff, "The quick brown fox jumped over the lazy dog", true, true , false)
	buff2 := glim.Rotate270(nw, nh, buff)
	nw, nh = nh, nw
	//glim.DumpBuff(buff,uint(nw),uint(nh))
	buff2 = glim.FlipUp(nw,nh,buff2)
	tt :=  glim.ImageToGFormatRGBA(nw,nh, buff2)
	w.Cmds().DrawImage(rect.Rect{0, 0, nw, nh}, tt)
}

func main() {
wnd := nucular.NewMasterWindow(0, "MyWindow", updatefn)
var theme nstyle.Theme = nstyle.DarkTheme
const scaling = 1.8
wnd.SetStyle(nstyle.FromTheme(theme, scaling))
wnd.Main()
}



func copyFormatter(inF *FormatParams) *FormatParams {
	out := NewFormatter()
	*out = *inF
	return out
}

//Draw some text into a 32bit RGBA byte array, wrapping where needed.  Supports all the options I need for a basic word processor, including vertical text, and different sized lines
//This was a bad idea.  Instead of all the if statements, we should just assume everything is left-to-right, top-to-bottom, and then rotate the entire block afterwards (we will also have to rotate the characters around their own center)
//Arabic will still need special code - better to separate into two completely different routines?
//Return the cursor position (number of characters from start of text) that is closest to the mouse cursor (cursorX, cursorY)
func RenderPara(f *glim.FormatParams, xpos, ypos, orig_xpos, orig_ypos, maxX, maxY, clientWidth, clientHeight, cursorX, cursorY int, u8Pix []uint8, text string, transparent bool, doDraw bool, showCursor bool) (int, int, int) {
	cursorDist := 9999999
	seekCursorPos := 0
	vert := f.Vertical
	orig_colour := f.Colour
	foreGround := f.Colour
	colSwitch := false
	if f.TailBuffer {
		//f.Cursor = len(text)
		//scrollToCursor(f, text)  //Use pageup function, once it is fast enough
	}
	//log.Printf("Cursor: %v\n", f.Cursor)
	letters := strings.Split(text, "")
	letters = append(letters, " ")
	orig_fontSize := f.FontSize
	defer func() {
		f.FontSize = orig_fontSize
		glim.SanityCheck(f, text)
	}()
	//xpos := orig_xpos
	//ypos := orig_ypos
	if vert {
		xpos = maxX
	}
	gx, gy := glim.GetGlyphSize(f.FontSize, text)
	//fmt.Printf("Chose position %v, maxX: %v\n", pos, maxX)
	pos := glim.MoveInBounds(glim.Vec2{xpos, ypos}, glim.Vec2{orig_xpos, orig_ypos}, glim.Vec2{maxX, maxY}, glim.Vec2{gx, gy}, glim.Vec2{0, 1}, glim.Vec2{-1, 0})
	xpos = pos.X
	ypos = pos.X
	maxHeight := 0
	letterWidth := 100
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
//			DrawCursor(xpos, ypos, maxHeight, clientWidth, u8Pix) //FIXME
		}
		if i >= len(letters)-1 {
			continue
		}
		//foreGround = orig_colour

		if unicode.IsSpace([]rune(v)[0]) {
			//if i>0 && letters[i-1] == " " {
			//f.Colour = &color.RGBA{255,0,0,255}
			//f.FontSize = f.FontSize*1.2
			////log.Printf("Oversize start for %v at %v\n", v, i)
			//} else {
			//f.Colour = &color.RGBA{1,1,1,255}
			//}
			colSwitch = !colSwitch
			if colSwitch {
				foreGround = &color.RGBA{255, 1, 1, 255}
			} else {
				foreGround = orig_colour
			}
		}
		if (i >= f.SelectStart) && (i <= f.SelectEnd) && (f.SelectStart != f.SelectEnd) {
			nf := glim.CopyFormatter(f)
			nf.SelectStart = -1
			nf.SelectEnd = -1
			nf.Colour = &color.RGBA{255, 1, 1, 255}
			/*if i-1<f.SelectStart {
			      _, xpos, ypos = RenderPara(nf, xpos, ypos, 0, 0, maxX, maxY, clientWidth, clientHeight, cursorX, cursorY, u8Pix, "{", transparent, doDraw, showCursor)
			  }
			  if i+1>f.SelectEnd {
			      _, xpos, ypos = RenderPara(nf, xpos, ypos, 0, 0, maxX, maxY, clientWidth, clientHeight, cursorX, cursorY, u8Pix, "}", transparent, doDraw, showCursor)
			  }*/

			//fmt.Printf("%v is between %v and %v\n", i , f.SelectStart, f.SelectEnd)
			foreGround = nf.Colour
		}
		//fmt.Printf("%v: %V\n", i , f)
		if (string(text[i]) == " ") || (string(text[i]) == "\n") {
			f.FontSize = orig_fontSize
			//log.Printf("Oversize end for %v at %v\n", v, i)
		}
		if string(text[i]) == "\n" {
			if vert {
				xpos = xpos - maxHeight
				ypos = orig_ypos
			} else {
				ypos = ypos + maxHeight
				xpos = orig_xpos
				if i > 0 && string(text[i-1]) != "\n" {
					maxHeight = 12 //FIXME
				}
			}
			f.Line++
			f.StartLinePos = i
			if f.Cursor == i && showCursor {
//FIXME				DrawCursor(xpos, ypos, maxHeight, clientWidth, u8Pix)
			}
		} else {
			if i >= f.FirstDrawnCharPos {
				ytweak := 0
				if wobblyMode {
					ytweak = int(math.Sin(float64(xpos)) * 5.0)
				}
				img, face := glim.DrawStringRGBA(f.FontSize, *foreGround, v, "f1.ttf")
				XmaX, YmaX := img.Bounds().Max.X, img.Bounds().Max.Y
				imgBytes := img.Pix
				//imgBytes := Rotate270(XmaX, YmaX, img.Pix)
				//XmaX, YmaX = YmaX, XmaX
				fa := *face
				glyph, _ := utf8.DecodeRuneInString(v)
				letterWidth_F, _ := fa.GlyphAdvance(glyph)
				letterWidth = glim.Fixed2int(letterWidth_F)
				//fuckedRect, _, _ := fa.GlyphBounds(glyph)
				//letterHeight := fixed2int(fuckedRect.Max.Y)
				letterHeight := glim.Fixed2int(fa.Metrics().Height)
				//letterWidth := XmaX
				//letterHeight = letterHeight

				if vert && (xpos < 0) {
					if vert {
						f.LastDrawnCharPos = i - 1
						return seekCursorPos, xpos, ypos
					} else {
						pos := glim.MoveInBounds(glim.Vec2{xpos, ypos}, glim.Vec2{orig_xpos, orig_ypos}, glim.Vec2{maxX, maxY}, glim.Vec2{gx, gy}, glim.Vec2{0, 1}, glim.Vec2{-1, 0})
						xpos = pos.X
						ypos = pos.Y
					}
				}
				if xpos+XmaX > maxX {
					if !vert {
						ypos = ypos + maxHeight
						maxHeight = 0
						xpos = orig_xpos
						f.Line++
						f.StartLinePos = i
					}
				}

				if (ypos+YmaX+ytweak+1 > maxY) || (ypos+ytweak < 0) {
					if vert {
						xpos = xpos - maxHeight
						maxHeight = 0
						ypos = orig_ypos
						f.Line++
						f.StartLinePos = i
					} else {
						f.LastDrawnCharPos = i - 1
						return seekCursorPos, xpos, ypos
					}
				}
				pos := glim.MoveInBounds(glim.Vec2{xpos, ypos}, glim.Vec2{orig_xpos, orig_ypos}, glim.Vec2{maxX, maxY}, glim.Vec2{XmaX, YmaX}, glim.Vec2{0, 1}, glim.Vec2{-1, 0})
				xpos = pos.X
				ypos = pos.Y

				if doDraw {
					//PasteImg(img, xpos, ypos + ytweak, u8Pix, transparent)
					//PasteBytes(XmaX, YmaX, imgBytes, xpos, ypos+ytweak, int(clientWidth), int(clientHeight), u8Pix, transparent)
					glim.PasteBytes(XmaX, YmaX, imgBytes, xpos, ypos+ytweak, int(clientWidth), int(clientHeight), u8Pix, true, false, true)
				}

				if f.Cursor == i && showCursor {
					//DrawCursor(xpos, ypos, maxHeight, clientWidth, u8Pix)
				}

				f.LastDrawnCharPos = i
				maxHeight = MaxI(maxHeight, letterHeight)

				if vert {
					ypos += maxHeight
				} else {
					xpos += letterWidth
				}
			}
		}
		d := (cursorX-xpos+letterWidth)*(cursorX-xpos+letterWidth) + (cursorY-ypos-maxHeight/2)*(cursorY-ypos-maxHeight/2)
		if d < cursorDist {
			cursorDist = d
			seekCursorPos = i
		}

	}
	glim.SanityCheck(f, text)
	return seekCursorPos, xpos, ypos
}

//Return the larger of two integers
func MaxI(a, b int) int {
	if a > b {
		return a
	}
	return b
}


//Holds all the configuration details for drawing a string into a texture.  This structure gets written to during the draw
type FormatParams struct {
	Colour            *color.RGBA //Text colour
	Line              int
	Cursor            int         //The cursor position, in characters from the start of the text
	SelectStart       int         //Start of the selection box, counted from the start of document
	SelectEnd         int         //End of the selection box, counted from the start of document
	StartLinePos      int         //Updated during render, holds the closest start of line, including soft line breaks
	FontSize          float64     //Fontsize, in points or something idfk
	FirstDrawnCharPos int         //The first character to draw on the screen.  Anything before this is ignored
	LastDrawnCharPos  int         //The last character that we were able to fit on the screen
	TailBuffer        bool        //Nothing for now
	Outline           bool        //Nothing for now
	Vertical          bool        //Draw texture vertically for Chinese/Japanese rendering
	SelectColour      *color.RGBA //Selection text colour
}

//Create a new text formatter, with useful default parameters
func NewFormatter() *FormatParams {
	return &FormatParams{&color.RGBA{5, 5, 5, 255}, 0, 0, 0, 0, 0, 22.0, 0, 0, false, true, false, &color.RGBA{255, 128, 128, 255}}
}

