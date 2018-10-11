package main

import (
	"github.com/donomii/glim"
	"github.com/donomii/nucular/rect"
	"log"

	//"image"
	"github.com/donomii/nucular"
	"image/color"
)

func DrawString(win *nucular.Window,x1,y1,x2,y2, w, h int, buff []byte) {
	img := glim.ImageToGFormatRGBA(w,h,buff)
win.Cmds().DrawImage(rect.Rect{50, 100, 200, 200}, img)
}

var demoText = ""
func updatefn(w *nucular.Window) {
	col := color.RGBA{255, 255, 255, 255}
	txtSize := 9.6
	if w.Input().Mouse.Buttons[1].Down {
		col = color.RGBA{255, 0, 0, 0}
		txtSize = 30
	}
	/*
	for _, v := range w.Input().Keyboard.Keys {
		log.Println("%+v", v.)
	}
	*/
	if w.Input().Keyboard.Text != "" {
		log.Println(w.Input().Keyboard.Text)
		demoText = demoText + w.Input().Keyboard.Text
	}
	w.Row(30).Dynamic(1)
	w.Label("Dynamic fixed column layout with generated position and size (LayoutRowDynamic):", "LC")
	w.Row(30).Dynamic(1)
	w.LabelColored("Hello", "LC", col)
	img, _ := glim.DrawStringRGBA(txtSize, col, "Hello again", "f1.ttf")
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
	f.FontSize = txtSize
	nw := 800
	nh := 500
	buff := make([]byte, nw*nh*4)
	
    glim.RenderPara(f , 10, 15, 0, 0, nw, nh, nw, nh, 1, 1, buff, demoText, true, true , false)
	buff2 := glim.Rotate270(nw, nh, buff)
	nw, nh = nh, nw
	//glim.DumpBuff(buff,uint(nw),uint(nh))
	buff2 = glim.FlipUp(nw,nh,buff2)
	tt :=  glim.ImageToGFormatRGBA(nw,nh, buff2)
	w.Cmds().DrawImage(rect.Rect{0, 0, nw, nh}, tt)
	log.Printf("%+v", w.Input())

}

func main() {
wnd := nucular.NewMasterWindow(0, "MyWindow", updatefn)
//var theme nstyle.Theme = nstyle.DarkTheme
//const scaling = 1.8
//wnd.SetStyle(nstyle.FromTheme(theme, scaling))
wnd.Main()
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

