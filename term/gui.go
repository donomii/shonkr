// gui.go
package main

import (
	"github.com/donomii/glim"
	"github.com/donomii/nuklear-templates"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"

	"fmt"

	"github.com/atotto/clipboard"

	"log"
	"os"
)

var foreColour, backColour *glim.RGBA
var mapTex *nktemplates.Texture
var mapTex1 *nktemplates.Texture
var lastEnterDown bool
var lastBackspaceDown bool

func drawmenu(ctx *nk.Context, state *State) {
	nk.NkMenubarBegin(ctx)

	menuItemWidth := float32(120)
	menuItemHeight := float32(200)
	barWidth := float32(95)
	/* menu #1 */
	nk.NkLayoutRowBegin(ctx, nk.Static, 25, 10)
	nk.NkLayoutRowPush(ctx, barWidth)

	if nk.NkMenuBeginLabel(ctx, "File", nk.TextLeft, nk.NkVec2(menuItemWidth, menuItemHeight)) > 0 {
		nk.NkLayoutRowDynamic(ctx, 25, 1)
		if nk.NkMenuItemLabel(ctx, "Save", nk.TextLeft) > 0 {
			dispatch("SAVE-FILE", ed)
		}
		if nk.NkMenuItemLabel(ctx, "Exit", nk.TextLeft) > 0 {
			os.Exit(0)
		}
		nk.NkMenuEnd(ctx)
	}

	if nk.NkMenuBeginLabel(ctx, "Edit", nk.TextLeft, nk.NkVec2(menuItemWidth, menuItemHeight)) > 0 {
		nk.NkLayoutRowDynamic(ctx, 25, 1)
		if nk.NkMenuItemLabel(ctx, "Paste", nk.TextLeft) > 0 {
			text, _ := clipboard.ReadAll()
			shellIn <- []byte(text)
			needsRedraw = true
		}
		if nk.NkMenuItemLabel(ctx, "Send Break", nk.TextLeft) > 0 {
			shellIn <- []byte{3}
			needsRedraw = true
		}
		nk.NkMenuEnd(ctx)
	}
	if nk.NkMenuBeginLabel(ctx, "Fonts", nk.TextLeft, nk.NkVec2(menuItemWidth, menuItemHeight)) > 0 {
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

	if nk.NkMenuBeginLabel(ctx, "Colours", nk.TextLeft, nk.NkVec2(menuItemWidth, menuItemHeight)) > 0 {
		nk.NkLayoutRowDynamic(ctx, 25, 1)
		if nk.NkMenuItemLabel(ctx, "White on Black", nk.TextLeft) > 0 {
			backColour = &glim.RGBA{255, 255, 255, 255}
			foreColour = &glim.RGBA{0, 0, 0, 255}
		}
		if nk.NkMenuItemLabel(ctx, "Black on White", nk.TextLeft) > 0 {
			foreColour = &glim.RGBA{255, 255, 255, 255}
			backColour = &glim.RGBA{0, 0, 0, 255}
		}
		nk.NkMenuEnd(ctx)
	}

	if nk.NkMenuBeginLabel(ctx, "Commands", nk.TextLeft, nk.NkVec2(menuItemWidth, menuItemHeight)) > 0 {
		nk.NkLayoutRowDynamic(ctx, 25, 1)
		if nk.NkMenuItemLabel(ctx, "Find and exec", nk.TextLeft) > 0 {
			shellIn <- []byte("find . -exec grep -s searchterm {} /dev/null \\;")
		}
		if nk.NkMenuItemLabel(ctx, "Recursive replace", nk.TextLeft) > 0 {
			shellIn <- []byte("find . -type f -name \"*.txt\" -exec sed -i'' -e 's/foo/bar/g' {} +")
		}
		nk.NkMenuEnd(ctx)
	}

	if nk.NkMenuBeginLabel(ctx, "Buffers", nk.TextLeft, nk.NkVec2(menuItemWidth, menuItemHeight)) > 0 {
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

		nk.NkCheckboxLabel(ctx, "check", &check)
		nk.NkMenuEnd(ctx)
	}

	nk.NkMenubarEnd(ctx)
}

func gfxMain(win *glfw.Window, ctx *nk.Context, state *State) {

	log.Println("Starting gfx")
	width, height := win.GetSize()
	log.Printf("glfw: window %vx%v", width, height)
	gl.Viewport(0, 0, int32(width-1), int32(height-1))
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

	if update > 0 {
		log.Println("Draw menu")
		drawmenu(ctx, state)
		log.Println("Draw editor")
		QuickFileEditor(ctx)

	}
	nk.NkEnd(ctx)
	log.Println("update complete")
	// Render
	bg := make([]float32, 4)
	nk.NkColorFv(bg, state.bgColor)
	width, height = win.GetSize()
	gl.Viewport(0, 0, int32(width), int32(height))

	//gl.Clear(gl.COLOR_BUFFER_BIT)
	//gl.ClearColor(0.0, 0.0, 0.0, 0.0) // Everything crashes if you move htis
	nk.NkPlatformRender(nk.AntiAliasingOn, maxVertexBuffer, maxElementBuffer)
	win.SwapBuffers()
	log.Println("Finished gfx")
}

func QuickFileEditor(ctx *nk.Context) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in QuickFileEditor", r)
		}
	}()

	nk.NkLayoutRowDynamic(ctx, float32(0), 2)
	{

		//log.Println("Check mouse")
		/*
			butts := ctx.Input().Mouse().GetButtons()

			mouseX, mouseY := int32(-1000), int32(-1000)

			for _, v := range butts {
				if *v.GetClicked() > 0 {
					mouseX, mouseY = ctx.Input().Mouse().Pos()

					//log.Println("Click at ", mouseX, mouseY)
				}
			}
		*/
		bounds := nk.NkWidgetBounds(ctx)
		left := int(*bounds.GetX())
		top := int(*bounds.GetY())
		nuHeight := 800
		log.Println("Starting row draw")
		nk.NkLayoutRowDynamic(ctx, float32(0), 1)
		{

			if ed != nil {
				log.Println("Starting editor draw")
				width := int(nk.NkWidgetWidth(ctx))

				size := width * nuHeight * 4
				log.Println("Clearing", size, "bytes(", width, "x", nuHeight, ")")
				for i := 0; i < size; i = i + 1 {
					pic[i] = ((*backColour)[0])
				}

				form = ed.ActiveBuffer.Formatter
				form.Colour = foreColour
				form.Colour = backColour
				form.Outline = true
				log.Println("Render paragraph")
				mouseX := 10
				mouseY := 10
				//screen :=
				//glim.RenderPara(ed.ActiveBuffer.Formatter,
				displayText := ed.ActiveBuffer.Data.Text
				if useAminal {
					displayText = aminalString(aminalTerm)
				}
				glim.RenderPara(ed.ActiveBuffer.Formatter,
					0, 0, 0, 0,
					width, nuHeight, width, nuHeight,
					int(mouseX)-left, int(mouseY)-top, pic, displayText,
					false, true, true)
				log.Println("Finished render paragraph")
				log.Println("Render image (", len(pic), " ", width, " ", nuHeight)
				doImage(ctx, pic, width, nuHeight)
			}
		}
	}

	log.Println("Finish terminal display")
}

func doImage(ctx *nk.Context, pic []uint8, width, nuHeight int) {
	log.Println("Rendering image")
	nk.NkLayoutRowDynamic(ctx, float32(nuHeight), 1)
	{
		var err error = nil
		log.Println("calling rawtex")

		mapTex1, err = nktemplates.RawTexture(glim.Uint8ToBytes(pic, picBytes), int32(width), int32(nuHeight), mapTex1)

		if err == nil {
			testim := nk.NkImageId(int32(mapTex1.Handle))
			log.Println("nk image")
			nk.NkImage(ctx, testim)
		} else {
			log.Println(err)
		}
	}
}
