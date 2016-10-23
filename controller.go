// +build !VR

package main

import (
    "github.com/atotto/clipboard"
    "io/ioutil"
    "fmt"
    "strings"
	"log"

	"golang.org/x/mobile/app"
	"golang.org/x/mobile/event/key"
	"golang.org/x/mobile/exp/sensor"
)

func handleSensor(e sensor.Event) {
	log.Println(e)
}

func scanToPrevLine (txt string, c int) int{
    letters := strings.Split(txt, "")
    x:=c
    for x= c-1; x>0 && x<len(txt) && !( letters[x-1]== "\n" && letters[x]!="\n"); x-- {}
    return x
}

func scanToNextLine (txt string, c int) int{
    letters := strings.Split(txt, "")
    x:=c
    for x= c+1; x>0 && x<len(txt) && !( letters[x-1]== "\n" && letters[x]!="\n"); x++ {}
    return x
}

func deleteLeft(t string, p int) string {
    if (p>0) {
        return strings.Join([]string{t[:p-1],t[p:]}, "")
    }
    return t
}

func saveFile(fname string, txt string ) {
    err := ioutil.WriteFile(fname, []byte(txt), 0644)
    check(err, "saving file")
}
        
    func check(e error, msg string) {
        if e != nil {
            log.Println("Error while ", msg, " : ", e)
        }
    }


var inputMode bool = false
func handleEvent(a app.App, i interface{}) {
	log.Println(i)
	switch e := a.Filter(i).(type) {
	case key.Event:
     switch e.Code {
            case key.CodeDeleteBackspace:
                if cursor > 0 {
                    txtBuff = deleteLeft(txtBuff, cursor)
                    cursor--
                }
            case key.CodeLeftArrow:
                cursor = cursor-1
            case key.CodeRightArrow:
                cursor = cursor+1
            case key.CodeUpArrow:
                cursor = scanToPrevLine(txtBuff, cursor)
            case key.CodeDownArrow:
                cursor = scanToNextLine(txtBuff, cursor)
   }
        if inputMode {
            switch e.Code {
            case key.CodeLeftShift:
            case key.CodeRightShift:
            case key.CodeReturnEnter:
                txtBuff = fmt.Sprintf("%s%s%s",txtBuff[:cursor], "\n",txtBuff[cursor:])
                cursor++
            default:
                switch e.Rune {
                    case '`':
                        inputMode = false
                    default:
                        txtBuff = fmt.Sprintf("%s%s%s",txtBuff[:cursor], string(e.Rune),txtBuff[cursor:])
                        cursor++
                }

            }
        } else {
            switch e.Code {
            case key.CodeA:
                cursor = cursor-1
            case key.CodeD:
                cursor = cursor+1
            case key.CodeQ:
                line = line +1
            case key.CodeE:
                line = line -1
            }
            switch e.Rune {
                case 'v':
                    text, _ := clipboard.ReadAll()
                        txtBuff = fmt.Sprintf("%s%s%s",txtBuff[:cursor], text,txtBuff[cursor:])
                case '~':
                    saveFile(fname, txtBuff)
                case 'i':
                    inputMode = true
                case 'w':
                    cursor = scanToPrevLine(txtBuff, cursor)
                case 's':
                    cursor = scanToNextLine(txtBuff, cursor)
                case 'W':
                    //Page up
                    start := searchBackPage(txtBuff, activeFormatter)
                    log.Println("New start at ", start)
                    activeFormatter.FirstDrawnCharPos = start
                    cursor = activeFormatter.FirstDrawnCharPos
                case 'S':
                    //page down
                    log.Println("Scanning to start of next page from ", activeFormatter.LastDrawnCharPos)
                    activeFormatter.FirstDrawnCharPos = scanToPrevLine(txtBuff,activeFormatter.LastDrawnCharPos)
                    cursor = activeFormatter.FirstDrawnCharPos
            }
        }

	}
}

