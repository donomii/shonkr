package main
import "github.com/donomii/nucular/rect"
//import "image"
import "github.com/donomii/nucular"
import "image/color"
import nstyle "github.com/donomii/nucular/style"
import "github.com/donomii/glim"
//import "github.com/disintegration/imaging"
//import "image/draw"

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

}

func main() {
wnd := nucular.NewMasterWindow(0, "MyWindow", updatefn)
var theme nstyle.Theme = nstyle.DarkTheme
const scaling = 1.8
wnd.SetStyle(nstyle.FromTheme(theme, scaling))
wnd.Main()
}
