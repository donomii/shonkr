// gui.go
package main

import (
	//"unsafe"
	//"io/ioutil"

	"github.com/donomii/glim"
	"github.com/donomii/nuklear-templates"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"

	//"text/scanner"

	"fmt"

	"github.com/atotto/clipboard"

	"log"
	"os"
	//"github.com/donomii/glim"
)

var DirFiles []string
var mapTex *nktemplates.Texture
var mapTex1 *nktemplates.Texture
var lastEnterDown bool
var lastBackspaceDown bool

func handleKeys(ctx *nk.Context) {
	keys := ctx.Input().Keyboard()
	//log.Printf("keys: %v\n", keys)

	var l *int32
	l = keys.GetTextLen()
	ll := *l
	if ll > 0 {
		if *(ctx.Input().GetKeyboard().GetTextLen()) > 0 {
			fmt.Printf("Input: %+v\n", ctx.Input())
			fmt.Printf("InputText: %+s\n", ctx.Input().GetKeyboard().GetText())
		}
	}

	if nk.NkInputIsKeyPressed(ctx.Input(), nk.KeyBackspace) > 0 {
		if lastBackspaceDown == false {
			dispatch("DELETE-LEFT", ed)
			go func() { shellIn <- []byte{127} }()
		}
		lastBackspaceDown = true
	} else {
		lastBackspaceDown = false
	}

	if nk.NkInputIsKeyPressed(ctx.Input(), nk.KeyTab) > 0 {
		go func() { shellIn <- []byte("\t") }()
	}
}
func drawmenu(ctx *nk.Context, state *State) {
	nk.NkMenubarBegin(ctx)

	/* menu #1 */
	nk.NkLayoutRowBegin(ctx, nk.Static, 25, 5)
	nk.NkLayoutRowPush(ctx, 45)

	if nk.NkMenuBeginLabel(ctx, "File", nk.TextLeft, nk.NkVec2(120, 200)) > 0 {
		nk.NkLayoutRowDynamic(ctx, 25, 1)
		if nk.NkMenuItemLabel(ctx, "Save", nk.TextLeft) > 0 {
			dispatch("SAVE-FILE", ed)
		}
		if nk.NkMenuItemLabel(ctx, "Exit", nk.TextLeft) > 0 {
			os.Exit(0)
		}
		nk.NkMenuEnd(ctx)
	}

	if nk.NkMenuBeginLabel(ctx, "Edit", nk.TextLeft, nk.NkVec2(120, 200)) > 0 {
		nk.NkLayoutRowDynamic(ctx, 25, 1)
		if nk.NkMenuItemLabel(ctx, "Paste", nk.TextLeft) > 0 {
			//dispatch("PASTE-FROM-CLIPBOARD", ed) //Adds it to the local buffer
			text, _ := clipboard.ReadAll()
			shellIn <- []byte(text)
		}
		if nk.NkMenuItemLabel(ctx, "Send Break", nk.TextLeft) > 0 {

			shellIn <- []byte{3}
		}
		nk.NkMenuEnd(ctx)
	}
	if nk.NkMenuBeginLabel(ctx, "Fonts", nk.TextLeft, nk.NkVec2(120, 200)) > 0 {
		//static size_t prog = 40;
		//static int slider = 10;
		check := int32(1)
		nk.NkLayoutRowDynamic(ctx, 25, 1)
		if nk.NkMenuItemLabel(ctx, "Text direction", nk.TextLeft) > 0 {
			dispatch("TOGGLE-VERTICAL-MODE", ed)
		}
		if nk.NkMenuItemLabel(ctx, "Increase font", nk.TextLeft) > 0 {
			dispatch("INCREASE-FONT", ed)
		}
		if nk.NkMenuItemLabel(ctx, "Decrease font", nk.TextLeft) > 0 {
			dispatch("DECREASE-FONT", ed)
		}
		if nk.NkMenuItemLabel(ctx, "8 point", nk.TextLeft) > 0 {
			SetFont(ed.ActiveBuffer, 8)
		}
		if nk.NkMenuItemLabel(ctx, "12 point", nk.TextLeft) > 0 {
			SetFont(ed.ActiveBuffer, 12)
		}
		if nk.NkMenuItemLabel(ctx, "20 point", nk.TextLeft) > 0 {
			SetFont(ed.ActiveBuffer, 20)
		}
		//if (nk.NkMenuItemLabel(ctx, "About", NK_TEXT_LEFT))
		//    show_app_about = nk_true;
		//			nk.NkProgress(ctx, &prog, 100, nk.Modifiable)
		//			nk.NkSliderInt(ctx, 0, &slider, 16, 1)
		nk.NkCheckboxLabel(ctx, "check", &check)
		nk.NkMenuEnd(ctx)
	}

	if nk.NkMenuBeginLabel(ctx, "Buffers", nk.TextLeft, nk.NkVec2(120, 200)) > 0 {
		//static size_t prog = 40;
		//static int slider = 10;
		check := int32(1)
		nk.NkLayoutRowDynamic(ctx, 25, 1)

		if nk.NkMenuItemLabel(ctx, "Clear Buffer", nk.TextLeft) > 0 {
			dispatch("CLEAR-BUFFER", ed)
			fmt.Println("Clear buffer")
		}

		if nk.NkMenuItemLabel(ctx, "Next Buffer", nk.TextLeft) > 0 {
			dispatch("NEXT-BUFFER", ed)
			fmt.Println("Next buffer")
		}
		if nk.NkMenuItemLabel(ctx, "Previous Buffer", nk.TextLeft) > 0 {
			dispatch("PREVIOUS-BUFFER", ed)
		}

		if nk.NkMenuItemLabel(ctx, "---------------", nk.TextLeft) > 0 {
		}

		for i, v := range ed.BufferList {
			if nk.NkMenuItemLabel(ctx, fmt.Sprintf("%v) %v", i, v.Data.FileName), nk.TextLeft) > 0 {
				ed.ActiveBuffer = ed.BufferList[i]
			}
		}

		//if (nk.NkMenuItemLabel(ctx, "About", NK_TEXT_LEFT))
		//    show_app_about = nk_true;
		//			nk.NkProgress(ctx, &prog, 100, nk.Modifiable)
		//			nk.NkSliderInt(ctx, 0, &slider, 16, 1)
		nk.NkCheckboxLabel(ctx, "check", &check)
		nk.NkMenuEnd(ctx)
	}

	nk.NkMenubarEnd(ctx)
}

func gfxMain(win *glfw.Window, ctx *nk.Context, state *State) {
	appName := "Shonkr"

	maxVertexBuffer := 512 * 1024
	maxElementBuffer := 128 * 1024

	nk.NkPlatformNewFrame()

	// Layout
	bounds := nk.NkRect(50, 50, 230, 250)
	update := nk.NkBegin(ctx, appName, bounds, nk.WindowBorder|nk.WindowMovable|nk.WindowScalable|nk.WindowMinimizable|nk.WindowTitle)

	col := nk.NewColor()
	col.SetRGBA(nk.Byte(255), nk.Byte(255), nk.Byte(255), nk.Byte(255))
	wbd := ctx.Style().Window().GetFixedBackground().GetData()
	wbd[0] = 255
	wbd[1] = 255
	wbd[2] = 255
	wbd[3] = 255
	wbg := ctx.Style().GetButton().GetTextBackground()
	wbg.SetRGBAi(255, 255, 255, 255)

	nk.NkWindowSetPosition(ctx, appName, nk.NkVec2(0, 0))
	nk.NkWindowSetSize(ctx, appName, nk.NkVec2(float32(winWidth), float32(winHeight)))

	handleKeys(ctx)

	if update > 0 {

		drawmenu(ctx, state)

		QuickFileEditor(ctx)

	}
	nk.NkEnd(ctx)

	// Render
	bg := make([]float32, 4)
	nk.NkColorFv(bg, state.bgColor)
	width, height := win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.ClearColor(bg[0], bg[1], bg[2], bg[3])
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
}

func QuickFileEditor(ctx *nk.Context) {

	nk.NkLayoutRowDynamic(ctx, float32(0), 2)
	{

		butts := ctx.Input().Mouse().GetButtons()

		mouseX, mouseY := int32(-1000), int32(-1000)

		for _, v := range butts {
			if *v.GetClicked() > 0 {
				mouseX, mouseY = ctx.Input().Mouse().Pos()

				log.Println("Click at ", mouseX, mouseY)
			}
		}
		bounds := nk.NkWidgetBounds(ctx)
		left := int(*bounds.GetX())
		top := int(*bounds.GetY())
		nuHeight := 8000
		nk.NkLayoutRowDynamic(ctx, float32(0), 1)
		{

			if ed != nil {
				width := int(nk.NkWidgetWidth(ctx))

				pic := make([]uint8, width*nuHeight*4)

				form.Colour = &glim.RGBA{255, 255, 255, 255}
				ed.ActiveBuffer.Formatter.Colour = &glim.RGBA{0, 0, 0, 255}
				ed.ActiveBuffer.Formatter.Outline = false
				newCursor, _, _ := glim.RenderPara(ed.ActiveBuffer.Formatter,
					0, 0, 0, 0,
					width, nuHeight, width, nuHeight,
					int(mouseX)-left, int(mouseY)-top, pic, ed.ActiveBuffer.Data.Text,
					true, true, true)
				for _, v := range butts {
					if *v.GetClicked() > 0 {
						form.Cursor = newCursor
						ed.ActiveBuffer.Formatter.Cursor = newCursor
					}
				}

				doImage(ctx, pic, width, nuHeight)
			}
		}

	}

}

func doImage(ctx *nk.Context, pic []uint8, width, nuHeight int) {
	nk.NkLayoutRowDynamic(ctx, float32(nuHeight), 1)
	{
		var err error = nil
		mapTex1, err = nktemplates.RawTexture(glim.Uint8ToBytes(pic), int32(width), int32(nuHeight), mapTex1)

		if err == nil {
			testim := nk.NkImageId(int32(mapTex1.Handle))
			nk.NkImage(ctx, testim)
		} else {
			log.Println(err)
		}
	}
}
