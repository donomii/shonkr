// gui.go
package main
import (
	"strings"

	//"unsafe"
	//"io/ioutil"
	"strconv"

	"github.com/donomii/glim"
	"github.com/donomii/nuklear-templates"

	"github.com/go-gl/gl/v3.2-core/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/golang-ui/nuklear/nk"

	//"text/scanner"

	"fmt"

	"log"
	"os"

	//"github.com/donomii/glim"
	"github.com/donomii/goof"
)

var DirFiles []string
var mapTex *nktemplates.Texture
var mapTex1 *nktemplates.Texture
var lastEnterDown bool
var lastBackspaceDown bool

func defaultMenu(ctx *nk.Context) {
	col := nk.NewColor()
	col.SetRGBA(nk.Byte(255), nk.Byte(255), nk.Byte(255), nk.Byte(255))
	nk.SetBackgroundColor(ctx, *col)
	if 0 < nk.NkButtonLabel(ctx, "---------") {
	}

	if 0 < nk.NkButtonLabel(ctx, "Edit Config") {
		LoadFileIfNotLoaded(ed, confFile)
		currentNode.Name = "File Manager"
	}

	if 0 < nk.NkButtonLabel(ctx, "Run command") {
		cmd := strings.Join(NodesToStringArray(currentThing[1:]), " ")
		result = goof.Command("cmd", []string{"/c", cmd})
		result = result + goof.Command("/bin/sh", []string{"-c", cmd})
	}

	if 0 < nk.NkButtonLabel(ctx, "Run command interactively") {
		goof.QCI(NodesToStringArray(currentThing[1:]))

	}
	if 0 < nk.NkButtonLabel(ctx, "Change directory") {
		path := strings.Join(NodesToStringArray(currentThing[1:]), "/")
		os.Chdir(path)
		currentNode = makeStartNode()
		currentThing = []*Node{currentNode}
	}

	if len(currentThing) > 1 {

		lastMenu := currentThing[len(currentThing)-2]

		if 0 < nk.NkButtonLabel(ctx, "Go back to "+lastMenu.Name) {
			if len(currentThing) > 1 {
				currentNode = currentThing[len(currentThing)-2]
				currentThing = currentThing[:len(currentThing)-1]
			}
		}
	}
	if 0 < nk.NkButtonLabel(ctx, "Exit") {

		fmt.Println(strings.Join(NodesToStringArray(currentThing), " ") + "\n")
		app.Stop()
		os.Exit(0)
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

		if nk.NkMenuItemLabel(ctx, "Next Buffer", nk.TextLeft) > 0 {
			dispatch("NEXT-BUFFER", ed)
			fmt.Println("NExt buffer")
		}
		if nk.NkMenuItemLabel(ctx, "Previous Buffer", nk.TextLeft) > 0 {
			dispatch("PREVIOUS-BUFFER", ed)
		}

		if nk.NkMenuItemLabel(ctx, "---------------", nk.TextLeft) > 0 {
		}

		//fmt.Printf("Bufferlist %+v\n", ed.BufferList)
		for i, v := range ed.BufferList {
			//fmt.Printf("Bufferlist %+v\n", v.Data)
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
	if nk.NkInputIsKeyPressed(ctx.Input(), nk.KeyEnter) > 0 {
		fmt.Printf("Enter: %+v\n", ctx.Input().GetKeyboard())
		if lastEnterDown == false {
			ed.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s", ed.ActiveBuffer.Data.Text[:ed.ActiveBuffer.Formatter.Cursor], "\n", ed.ActiveBuffer.Data.Text[ed.ActiveBuffer.Formatter.Cursor:])
			ed.ActiveBuffer.Formatter.Cursor++
		}
		lastEnterDown = true
	} else {
		lastEnterDown = false
	}

	if nk.NkInputIsKeyPressed(ctx.Input(), nk.KeyBackspace) > 0 {
		fmt.Printf("Back: %+v\n", ctx.Input().GetKeyboard())
		if lastBackspaceDown == false {
			dispatch("DELETE-LEFT", ed)
		}
		lastBackspaceDown = true
	} else {
		lastBackspaceDown = false
	}

	if nk.NkInputIsKeyPressed(ctx.Input(), nk.KeyLeft) > 0 {
		dispatch("PREVIOUS-CHARACTER", ed)
	}

	if nk.NkInputIsKeyPressed(ctx.Input(), nk.KeyRight) > 0 {
		dispatch("NEXT-CHARACTER", ed)
	}

	if nk.NkInputIsKeyPressed(ctx.Input(), nk.KeyDown) > 0 {
		dispatch("NEXT-LINE", ed)
	}

	if nk.NkInputIsKeyPressed(ctx.Input(), nk.KeyUp) > 0 {
		dispatch("PREVIOUS-LINE", ed)
	}

	if update > 0 {

		drawmenu(ctx, state)

		nk.NkLayoutRowDynamic(ctx, 20, 3)
		{
			nk.NkLabel(ctx, strings.Join(NodesToStringArray(currentThing), " > "), nk.TextLeft)
			if 0 < nk.NkButtonLabel(ctx, "Undo") {
				if len(currentThing) > 1 {
					currentNode = currentThing[len(currentThing)-2]
					currentThing = currentThing[:len(currentThing)-1]
				}
			}
			if 0 < nk.NkButtonLabel(ctx, "Go Back") {
				if len(currentThing) > 1 {
					currentNode = currentThing[len(currentThing)-2]
				}
			}
		}

		if currentNode.Name == "File Manager" {
			QuickFileEditor(ctx)
		} else {
			ButtonBox(ctx)
		}

		nk.NkLayoutRowDynamic(ctx, 20, 3)
		{

			if 0 < nk.NkButtonLabel(ctx, "Run interactive") {

			}
		}
		nk.NkLayoutRowDynamic(ctx, 25, 1)
		{

			//pf := nk.NewPluginFilterRef(unsafe.Pointer(&nk.NkFilterDefault))

		}

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

func ButtonBox(ctx *nk.Context) {
	nk.NkLayoutRowDynamic(ctx, 400, 2)
	{
		nk.NkGroupBegin(ctx, "Group 1", nk.WindowBorder)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		{
			for _, vv := range currentNode.SubNodes {
				//node := vv.SubNodes[i]
				name := vv.Name
				command := vv.Command
				v := vv
				if nk.NkButtonLabel(ctx, name) > 0 {
					fmt.Println("Data:", vv.Data)
					result = vv.Data
					if !strings.HasPrefix(command, "!") && !strings.HasPrefix(command, "&") {
						currentThing = append(currentThing, v)
						currentNode = v
					}
				}
			}

			defaultMenu(ctx)
		}
		nk.NkGroupEnd(ctx)

		nk.NkGroupBegin(ctx, "Group 2", nk.WindowBorder)
		nk.NkLayoutRowDynamic(ctx, 10, 1)
		{
			//Control the display
			nk.NkLayoutRowDynamic(ctx, 20, 3)
			{

				if 0 < nk.NkButtonLabel(ctx, "None") {
					displaySplit = "None"
				}

				if 0 < nk.NkButtonLabel(ctx, "Spaces") {
					displaySplit = "Spaces"
				}

				if 0 < nk.NkButtonLabel(ctx, "Tabs") {
					displaySplit = "Tabs"
				}

				if displaySplit == "None" {
					nk.NkLayoutRowDynamic(ctx, 10, 1)
					{
						results := strings.Split(result, "\n")
						for _, v := range results {
							//nk.NkLabel(ctx, v, nk.WindowBorder)
							if nk.NkButtonLabel(ctx, v) > 0 {
								n := makeNodeShort(v, []*Node{})
								currentThing = append(currentThing, n)

							}
						}
					}
				}

				if displaySplit == "Spaces" {
					nk.NkLayoutRowDynamic(ctx, 10, 5)
					{
						results := strings.Split(result, "\n")
						for _, line := range results {
							bits := strings.Split(line, " ")
							for i := 0; i < 5; i++ {
								label := ""
								if i < len(bits) {
									label = bits[i]
								} else {
									label = ""
								}
								//nk.NkLabel(ctx, v, nk.WindowBorder)
								if nk.NkButtonLabel(ctx, label) > 0 {
									n := makeNodeShort(label, []*Node{})
									currentThing = append(currentThing, n)

								}
							}
						}
					}
				}

			}

		}
		nk.NkGroupEnd(ctx)
	}

}

func QuickFileEditor(ctx *nk.Context) {

	nk.NkLayoutRowDynamic(ctx, float32(winHeight), 2)
	{
		nk.NkGroupBegin(ctx, "Group 1", nk.WindowBorder)
		nk.NkLayoutRowDynamic(ctx, 20, 1)
		{

			for _, vv := range DirFiles {
				//fmt.Println("File ",i, ":", vv)
				if nk.NkButtonLabel(ctx, vv) > 0 {
					fmt.Println("Loading file ", vv)
					LoadFileIfNotLoaded(ed, vv)
					//var err error
					//EditBytes, err = ioutil.ReadFile(vv)
					//AddActiveBuffer(ed, string(EditBytes), vv)
					//log.Println(err)
				}

			}
		}

		w := 200
		h := 200
		p1 := make([]uint8, w*h*4)
		//Want to display status bar here
		//fmt.Printf("Bufferlist %+v\n", ed.BufferList)
		buff := ed.StatusBuffer
		glim.RenderPara(buff.Formatter, 0, 0, 0, 0, w, h, w, h, 0, 0, p1, buff.Data.Text, true, true, true)
		doImage(ctx, p1, w, h)
		nk.NkGroupEnd(ctx)

		nk.NkGroupBegin(ctx, "Group 2", nk.WindowBorder)

		//nk.NkLayoutRowStatic(ctx, 100, 100, 3)
		//nk.NkLayoutRowDynamic(ctx, float32(winHeight), 1)
		height := 500
		butts := ctx.Input().Mouse().GetButtons()
		keys := ctx.Input().Keyboard()

		text := keys.GetText()
		var l *int32
		l = keys.GetTextLen()
		ll := *l
		if ll > 0 {
			if *(ctx.Input().GetKeyboard().GetTextLen()) > 0 {
				fmt.Printf("%+v\n", ctx.Input())
				fmt.Printf("%+s\n", ctx.Input().GetKeyboard().GetText())
			}
			s := fmt.Sprintf("\"%vu%04x\"", `\`, int(text[0]))
			s2, _ := strconv.Unquote(s)
			/*log.Println(err)
			log.Printf("Text: %v, %v\n", s, s2)
			newBytes := append(EditBytes[:form.Cursor], []byte(s2)...)
			newBytes = append(newBytes, EditBytes[form.Cursor:]...)
			form.Cursor++
			*/
			if ed.ActiveBuffer.Formatter.Cursor < 0 {
				ed.ActiveBuffer.Formatter.Cursor = 0
			}

			fmt.Printf("Inserting at %v, length %v\n", ed.ActiveBuffer.Formatter.Cursor, len(ed.ActiveBuffer.Data.Text))
			ed.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s", ed.ActiveBuffer.Data.Text[:ed.ActiveBuffer.Formatter.Cursor], fmt.Sprintf("%v", s2), ed.ActiveBuffer.Data.Text[ed.ActiveBuffer.Formatter.Cursor:])
			ed.ActiveBuffer.Formatter.Cursor++
		}
		mouseX, mouseY := int32(-1000), int32(-1000)

		for _, v := range butts {
			if *v.GetClicked() > 0 {
				//fmt.Printf("%+v\n", ctx.Input())
				mouseX, mouseY = ctx.Input().Mouse().Pos()

				log.Println("Click at ", mouseX, mouseY)
			}
		}
		bounds := nk.NkWidgetBounds(ctx)
		left := int(*bounds.GetX())
		top := int(*bounds.GetY())
		nuHeight := height - top
		nk.NkLayoutRowDynamic(ctx, float32(nuHeight), 1)
		{
			//if EditBytes != nil {
			if ed != nil {
				nkwidth := nk.NkWidgetWidth(ctx)
				width := int(nkwidth)

				//fmt.Println("Width:", width)
				//var lenStr = int32(len(EditBytes))
				//nk.NkEditString(ctx, nk.EditMultiline|nk.EditAlwaysInsertMode, EditBytes, &lenStr, 512, nk.NkFilterAscii) FIXME
				//nk.NkLabelWrap(ctx, string(EditBytes))
				pic := make([]uint8, width*nuHeight*4)

				form.Colour = &glim.RGBA{255, 255, 255, 255}
				//form.Cursor = 20
				form.FontSize = 12
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

				//pic, width, height := glim.GFormatToImage(im, nil, width, height)
				//gl.DeleteTextures(testim)
				//t, err := nktemplates.LoadImageFile(fmt.Sprintf("%v/progress%05v.png", output, fnum), width, height)
				//t := nktemplates.LoadImageData(globalPic, width, height)
<<<<<<< HEAD
				mapTex, _ = nktemplates.RawTexture(glim.Uint8ToBytes(pic, nil),
=======
				mapTex, _ = nktemplates.RawTexture(glim.Uint8ToBytes(pic,nil),
>>>>>>> 5157902777409f66343165956faea20eed2dd41b
					int32(width), int32(nuHeight), mapTex)
				var err error = nil
				if err == nil {
					testim := nk.NkImageId(int32(mapTex.Handle))
					//nk.NkLayoutRowStatic(ctx, 400, 400, 1)
					//{
					//log.Println("Drawing image")

					/*
						if nk.NkButtonImage(ctx, testim) > 0 {
							ed.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s", ed.ActiveBuffer.Data.Text[:ed.ActiveBuffer.Formatter.Cursor], "\n", ed.ActiveBuffer.Data.Text[ed.ActiveBuffer.Formatter.Cursor:])
							ed.ActiveBuffer.Formatter.Cursor++
						}
					*/
					nk.NkImage(ctx, testim)
					//}
				} else {
					log.Println(err)
				}
			}
		}
		nk.NkGroupEnd(ctx)
	}

}

func doImage(ctx *nk.Context, pic []uint8, width, nuHeight int) {
	nk.NkLayoutRowDynamic(ctx, float32(nuHeight), 1)
	{
<<<<<<< HEAD
		mapTex1, _ = nktemplates.RawTexture(glim.Uint8ToBytes(pic,nil), int32(width), int32(nuHeight), mapTex1)
=======
		mapTex1, _ = nktemplates.RawTexture(glim.Uint8ToBytes(pic, nil), int32(width), int32(nuHeight), mapTex1)
>>>>>>> 5157902777409f66343165956faea20eed2dd41b
		var err error = nil
		if err == nil {
			testim := nk.NkImageId(int32(mapTex1.Handle))
			nk.NkImage(ctx, testim)
		}
	}
}
