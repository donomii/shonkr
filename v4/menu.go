package main

import (
  "os"
   "github.com/donomii/menu"
 )
var rootMenu *menu.Node

func addMenuItem(currentMenu *menu.Node, menuText string, f func()) {
	item := menu.MakeNodeShort(menuText, nil)
	item.Function = f
	menu.AppendNode(currentMenu, item)
}

func setup_menu(){ 
	currentMenu = menu.MakeNodeShort("Main Menu", nil)
	rootMenu = currentMenu
	addMenuItem(currentMenu, "Go to start", func() { dispatch("START-OF-FILE", ed); update = true; mode = "searching" })
	addMenuItem(currentMenu, "Go to end", func() { dispatch("END-OF-FILE", ed); update = true; mode = "searching" })
	addMenuItem(currentMenu, "Increase Font", func() { dispatch("INCREASE-FONT", ed); update = true; mode = "searching" })

	addMenuItem(currentMenu, "Decrease Font", func() { dispatch("DECREASE-FONT", ed); update = true; mode = "searching" })

	addMenuItem(currentMenu, "Vertical Mode", func() { dispatch("VERTICAL-MODE", ed); update = true; mode = "searching" })

	addMenuItem(currentMenu, "Horizontal Mode", func() { dispatch("HORIZONTAL-MODE", ed); update = true; mode = "searching" })

	addMenuItem(currentMenu, "Save file", func() { dispatch("SAVE-FILE", ed); update = true; mode = "searching" })

	addMenuItem(currentMenu, "Exit", func() { os.Exit(0)})

	item := menu.MakeNodeShort("Switch Buffer", nil)
	item.Function = func() {
		buffMenu := menu.MakeNodeShort("Buffer Menu", nil)
		for i, v := range ed.BufferList {
			ii := i
			addMenuItem(buffMenu, v.Data.FileName, func() { ed.ActiveBuffer = ed.BufferList[ii]; mode = "searching"; currentMenu = rootMenu; update = true })
		}
		currentMenu = buffMenu
	}
	menu.AppendNode(currentMenu, item)
}
