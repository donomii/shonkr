// +build !VR

package main

import (
    "github.com/donomii/glim"
    "io"
    "golang.org/x/crypto/ssh/agent"
    "net"
    "os"
    //"github.com/mitchellh/go-homedir"
    "golang.org/x/crypto/ssh"
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

func Log2Buff(s string) {
	gc.StatusBuffer.Data.Text = s
}
func SearchBackPage(txtBuf string, orig_f *glim.FormatParams, screenWidth, screenHeight int)  int {
	input := *orig_f
	x := input.StartLinePos
	newLastDrawn := input.LastDrawnCharPos
	for x = input.Cursor; x > 0 && input.FirstDrawnCharPos < newLastDrawn; x = scanToPrevLine(txtBuf, x) {
		f := input
		f.FirstDrawnCharPos = x
        
		glim.RenderPara(&f, 0, 0, 0, 0, screenWidth/2, screenHeight, screenWidth/2, screenHeight, 0,0, nil, txtBuf, false, false, false)
		newLastDrawn = f.LastDrawnCharPos
	}
	return x
}



func DumpBuffer(b *Buffer) {
	Log2Buff(fmt.Sprintf(`
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



func scanToPrevPara (txt string, c int) int{
    log.Println("To Previous Line")
    letters := strings.Split(txt, "")
    x:=c
    for x= c-1; x>1 && x<len(txt) && !( letters[x-1]== "\n" && letters[x]!="\n"); x-- {}
    return x
}

func scanToPrevLine (txt string, c int) int{
    log.Println("To Previous Line")
    letters := strings.Split(txt, "")
    x:=c
    for x=c-1; x>1 && x<len(txt) && !( letters[x-1]== "\n"); x-- {}
    return x
}

func is_space (l string) bool{
    if (
        (l == " ") ||
        (l == "\n") ||
        (l == "\t")) {
        return true
    }
    return false
}

func SOL (txt string, c int) int{
    if (c==0) {
        return c
    }
    letters := strings.Split(txt, "")
    if (letters[c-1] == "\n") {
        return c
    }
    s := scanToPrevLine(txt, c)
    return s
}
func SOT (txt string, c int) int{  //Start of Text
    s := SOL(txt, c)
    letters := strings.Split(txt, "")
    x:=c
    for x=s; x>1 && x<len(txt) && (is_space(letters[x])); x++ {}
    return x
}

func scanToNextPara (txt string, c int) int{
    letters := strings.Split(txt, "")
    x:=c
    for x= c+1; x>1 && x<len(txt) && !( letters[x-1]== "\n" && letters[x]!="\n"); x++ {}
    return x
}

func scanToNextLine (txt string, c int) int{
    letters := strings.Split(txt, "")
    x:=c
    for x= c+1; x>1 && x<len(txt) && !( letters[x-1]== "\n"); x++ {}
    if x == len(txt) {
        return c
    } else {
        return x
    }
}


func scanToEndOfLine (txt string, c int) int{
    log.Println("To EOL")
    letters := strings.Split(txt, "")
    x:=c
    for x= c+1; x>0 && x<len(txt) && !( letters[x]== "\n"); x++ {}
    return x
}

func deleteLeft(t string, p int) string {
    log.Println("Delete left")
    if (p>0) {
        return strings.Join([]string{t[:p-1],t[p:]}, "")
    }
    return t
}

func saveFile(fname string, txt string ) {
    Log2Buff(fmt.Sprintf("Saving: %v",fname))
    err := ioutil.WriteFile(fname, []byte(txt), 0644)
    check(err, "saving file")
    Log2Buff(fmt.Sprintf("File saved: %v",fname))
}

func check(e error, msg string) {
    if e != nil {
        log.Println("Error while ", msg, " : ", e)
    }
}

func processPort(r io.Reader) {
    for {
        buf := make([]byte, 1)
        if _, err := io.ReadAtLeast(r, buf, 5); err != nil {
            //log.Fatal(err)
        }
        activeBufferAppend(string(buf))
        gc.ActiveBuffer.Formatter.Cursor = len(gc.ActiveBuffer.Data.Text)
    }
}

func buffAppend ( buffId int, txt string ) {
        gc.BufferList[1].Data.Text = strings.Join([]string{gc.BufferList[1].Data.Text,txt}, "")
}

func activeBufferAppend ( txt string ) {
        gc.ActiveBuffer.Data.Text = strings.Join([]string{gc.ActiveBuffer.Data.Text,txt}, "")
}

func SSHAgent() ssh.AuthMethod {
    if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
        return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
    }
    return nil
}

func startSshConn(buffId int, user, password, serverAndPort string) {
                    activeBufferAppend("Starting ssh connection\n")
                    config := &ssh.ClientConfig{
                        User: user,
                        Auth: []ssh.AuthMethod{
                            // Use the PublicKeys method for remote authentication.
                            SSHAgent(),
                        },
                    }

                    // Dial your ssh server.
                    activeBufferAppend(fmt.Sprintf("Connecting to ssh server as user %v: ", user))
                    activeBufferAppend(serverAndPort)
                    conn, err := ssh.Dial("tcp", serverAndPort, config)
                    if err != nil {
                        log.Fatal("unable to connect: ", err)
                    }

                    session, err := conn.NewSession()
                    if err != nil {
                        //return fmt.Errorf("Failed to create session: %s", err)
                    }

                    stdin, err := session.StdinPipe()
                    if err != nil {
                        //return fmt.Errorf("Unable to setup stdin for session: %v", err)
                    }
                    go io.Copy(stdin, os.Stdin)

                    stdout, err := session.StdoutPipe()
                    if err != nil {
                        //return fmt.Errorf("Unable to setup stdout for session: %v", err)
                    }
                    //go io.Copy(os.Stdout, stdout)
                    go processPort(stdout)
                    dispatch("TAIL", gc)

                    stderr, err := session.StderrPipe()
                    if err != nil {
                        //return fmt.Errorf("Unable to setup stderr for session: %v", err)
                    }
                    go io.Copy(os.Stderr, stderr)

                    err = session.Run("dude tail-all-logs")
                    defer conn.Close()
}

func pageDown(buf *Buffer) {
    log.Println("Scanning to start of next page from ", buf.Formatter.LastDrawnCharPos)
    buf.Formatter.FirstDrawnCharPos = scanToPrevLine(buf.Data.Text,buf.Formatter.LastDrawnCharPos)
    buf.Formatter.Cursor = buf.Formatter.FirstDrawnCharPos
}

func scrollToCursor(buf *Buffer){
    buf.Formatter.FirstDrawnCharPos = buf.Formatter.Cursor
}

func exciseSelection(buf *Buffer) {
    if buf.Formatter.SelectStart >= 0 && buf.Formatter.SelectStart < len(buf.Data.Text) {
        if buf.Formatter.SelectEnd > 0 && buf.Formatter.SelectEnd < len(buf.Data.Text) {
            log.Println("Clipping from ", buf.Formatter.SelectStart, " to ", buf.Formatter.SelectEnd)
            buf.Data.Text = fmt.Sprintf("%s%s",
            buf.Data.Text[:buf.Formatter.SelectStart],
            buf.Data.Text[buf.Formatter.SelectEnd+1:])
            buf.Formatter.SelectStart = 0
            buf.Formatter.SelectEnd = 0
        }
    }
}

func reduceFont(buf *Buffer) {
  buf.Formatter.FontSize -= 1
  glim.ClearAllCaches()

}

func increaseFont(buf *Buffer) {
  buf.Formatter.FontSize += 1
  glim.ClearAllCaches()
}

func doPageDown(buf *Buffer) {
    pageDown(gc.ActiveBuffer)
}

func previousCharacter(buf *Buffer) {
    buf.Formatter.Cursor = buf.Formatter.Cursor-1
}

func nextBuffer(gc GlobalConfig) {
                    gc.ActiveBufferId ++
                    if gc.ActiveBufferId>len(gc.BufferList)-1 {
                        gc.ActiveBufferId = 0
                    }
                    gc.ActiveBuffer = gc.BufferList[gc.ActiveBufferId]
                    log.Printf("Next buffer: %v", gc.ActiveBufferId)
}

func toggleVerticalMode(gc GlobalConfig) {
    if gc.ActiveBuffer.Formatter.Vertical {
        dispatch("HORIZONTAL-MODE", gc)
    } else {
        dispatch("VERTICAL-MODE", gc)
    }
}


func pasteFromClipBoard(buf *Buffer) {
    text, _ := clipboard.ReadAll()
    dispatch("EXCISE-SELECTION", gc)

if gc.ActiveBuffer.Formatter.Cursor < 0 {
    gc.ActiveBuffer.Formatter.Cursor = 0
}
   
    gc.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s",gc.ActiveBuffer.Data.Text[:gc.ActiveBuffer.Formatter.Cursor], text,gc.ActiveBuffer.Data.Text[gc.ActiveBuffer.Formatter.Cursor:])
}

func dispatch (command string, gc GlobalConfig) {
    switch command {
		case "WHEEL-UP":
		gc.ActiveBuffer.Formatter.Cursor = scanToPrevLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
		case "WHEEL-DOWN":
		gc.ActiveBuffer.Formatter.Cursor = scanToNextLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
        case "EXCISE-SELECTION":
            exciseSelection(gc.ActiveBuffer)
        case "REDUCE-FONT":
            reduceFont(gc.ActiveBuffer)
        case "INCREASE-FONT":
            increaseFont(gc.ActiveBuffer)
        case "PAGEDOWN":
            doPageDown(gc.ActiveBuffer)
        case "PAGEUP":
            pageUp(gc.ActiveBuffer, screenWidth, screenHeight)
        case "PREVIOUS-CHARACTER":
            previousCharacter(gc.ActiveBuffer)
        case "NEXT-CHARACTER":
            gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.Cursor+1
        case "PREVIOUS-LINE":
            gc.ActiveBuffer.Formatter.Cursor = scanToPrevLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
        case "NEXT-LINE":
            gc.ActiveBuffer.Formatter.Cursor = scanToNextLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
        case "NEXT-BUFFER":
            nextBuffer(gc)
        case "INPUT-MODE":
            gc.ActiveBuffer.InputMode = true
        case "START-OF-LINE":
           gc.ActiveBuffer.Formatter.Cursor = SOL(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
        case "HORIZONTAL-MODE":
            gc.ActiveBuffer.Formatter.Vertical = false
        case "VERTICAL-MODE":
            gc.ActiveBuffer.Formatter.Vertical = true
        case "TOGGLE-VERTICAL-MODE":
            toggleVerticalMode(gc)
        case "PASTE-FROM-CLIPBOARD":
            pasteFromClipBoard(gc.ActiveBuffer)
        case "COPY-TO-CLIPBOARD":
            clipboard.WriteAll(gc.ActiveBuffer.Data.Text[gc.ActiveBuffer.Formatter.SelectStart:gc.ActiveBuffer.Formatter.SelectEnd+1])
        case "SAVE-FILE":
            saveFile(gc.ActiveBuffer.Data.FileName, gc.ActiveBuffer.Data.Text)
        case "SEEK-EOL":
           gc.ActiveBuffer.Formatter.Cursor = scanToEndOfLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
        case "END-OF-LINE":
           gc.ActiveBuffer.Formatter.Cursor = scanToEndOfLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
        case "TAIL":
            gc.ActiveBuffer.Formatter.TailBuffer = true
        case "START-OF-TEXT-ON-LINE":
           gc.ActiveBuffer.Formatter.Cursor = SOT(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
    }
}


func pageUp(buf *Buffer, w,h int) {
    log.Println("Page up")
    start := SearchBackPage(buf.Data.Text, buf.Formatter, w, h)
    log.Println("New start at ", start)
    buf.Formatter.FirstDrawnCharPos = start
    buf.Formatter.Cursor = buf.Formatter.FirstDrawnCharPos
}

func handleEvent(a app.App, i interface{}) {
	log.Println(i)
    DumpBuffer(gc.ActiveBuffer)
	switch e := a.Filter(i).(type) {
	case key.Event:
     switch e.Code {
            case key.CodeDeleteBackspace:
                if gc.ActiveBuffer.Formatter.Cursor > 0 {
                    gc.ActiveBuffer.Data.Text = deleteLeft(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
                    gc.ActiveBuffer.Formatter.Cursor--
                }
            case key.CodeHome:
                gc.ActiveBuffer.Formatter.Cursor = SOL(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
            case key.CodeEnd:
                   dispatch("SEEK-EOL", gc)
            case key.CodeLeftArrow:
                   dispatch("PREVIOUS-CHARACTER", gc)
            case key.CodeRightArrow:
                   dispatch("NEXT-CHARACTER", gc)
            case key.CodeUpArrow:
                   dispatch("PREVIOUS-LINE", gc)
            case key.CodeDownArrow:
                   dispatch("NEXT-LINE", gc)
            case key.CodePageDown:
                dispatch("PAGEDOWN", gc)
            case key.CodePageUp:
                dispatch("PAGEUP", gc)
            case key.CodeA:
                if e.Modifiers >0 {
                    gc.ActiveBuffer.Formatter.SelectStart = 0
                    gc.ActiveBuffer.Formatter.SelectEnd = len(gc.ActiveBuffer.Data.Text) -1
                    return
                }
            case key.CodeC:
                if e.Modifiers >0 {
                    dispatch("COPY-TO-CLIPBOARD", gc)
                    return
                }
            case key.CodeX:
                if e.Modifiers >0 {
                    dispatch("COPY-TO-CLIPBOARD", gc)
                    dispatch("EXCISE-SELECTION", gc)
                    gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.SelectStart 
                    gc.ActiveBuffer.Formatter.SelectStart = -1
                    gc.ActiveBuffer.Formatter.SelectEnd = -1
                    return
                }
            case key.CodeV:
                if e.Modifiers >0 {
                    dispatch("EXCISE-SELECTION", gc)
                    dispatch("PASTE-FROM-CLIPBOARD", gc)
                }
            default:
       if gc.ActiveBuffer.InputMode {
            switch e.Code {
            case key.CodeLeftShift:
            case key.CodeRightShift:
            case key.CodeReturnEnter:
                gc.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s",gc.ActiveBuffer.Data.Text[:gc.ActiveBuffer.Formatter.Cursor], "\n",gc.ActiveBuffer.Data.Text[gc.ActiveBuffer.Formatter.Cursor:])
                gc.ActiveBuffer.Formatter.Cursor++
            default:
                switch e.Rune {
                    case '`':
                        gc.ActiveBuffer.InputMode = false
                    default:
                        if gc.ActiveBuffer.Formatter.SelectEnd > 0 {
                            dispatch("EXCISE-SELECTION", gc)
                        }
						if gc.ActiveBuffer.Formatter.Cursor < 0 {
						   gc.ActiveBuffer.Formatter.Cursor = 0
						}
                        fmt.Printf("Inserting at %v, length %v\n", gc.ActiveBuffer.Formatter.Cursor, len(gc.ActiveBuffer.Data.Text))
                        gc.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s",gc.ActiveBuffer.Data.Text[:gc.ActiveBuffer.Formatter.Cursor], string(e.Rune),gc.ActiveBuffer.Data.Text[gc.ActiveBuffer.Formatter.Cursor:])
                        gc.ActiveBuffer.Formatter.Cursor++
                }

            }
        } else {
            switch e.Code {
            case key.CodeX:
                if e.Modifiers >0 {
                    dispatch("EXCISE-SELECTION", gc)
                }
                
            case key.CodeA:
                if e.Modifiers >0 {
                    gc.ActiveBuffer.Formatter.SelectStart = 0
                    gc.ActiveBuffer.Formatter.SelectEnd = len(gc.ActiveBuffer.Data.Text)
                }
                gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.Cursor-1
            case key.CodeD:
                gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.Cursor+1
            case key.CodeQ:
                gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.Cursor +1
            case key.CodeE:
                gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.Cursor -1
            }
            switch e.Rune {
                case 'L':
                    go startSshConn(1, "", "", "")
                case 'N':
                    dispatch("NEXT-BUFFER", gc)
                case 'p':
                    dispatch("PASTE-FROM-CLIPBOARD", gc)
                case 'y':
                    dispatch("COPY-TO-CLIPBOARD", gc)
                case '~':
                    dispatch("SAVE-FILE", gc)
                case 'i':
                   dispatch("INPUT-MODE", gc)
                case '0':
                   dispatch("START-OF-LINE", gc)
                case '^':
                    dispatch("START-OF-TEXT-ON-LINE", gc)
                case '$':
                   dispatch("END-OF-LINE", gc)
                case 'A':
                   dispatch("END-OF-LINE", gc)
                   dispatch("INPUT-MODE", gc)
                case 'a':
                   gc.ActiveBuffer.Formatter.Cursor++
                   dispatch("INPUT-MODE", gc)
                case 'k':
                    gc.ActiveBuffer.Formatter.Cursor = scanToPrevLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
                case 'j':
                    gc.ActiveBuffer.Formatter.Cursor = scanToNextLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
                case 'l':
                   dispatch("NEXT-CHARACTER", gc)
                case 'h':
                   dispatch("PREVIOUS-CHARACTER", gc)
                case 'T':
                   dispatch("TAIL", gc)
                case 'W':
                    if gc.ActiveBuffer.Formatter.Outline {
                        gc.ActiveBuffer.Formatter.Outline = false
                    } else {
                        gc.ActiveBuffer.Formatter.Outline = true
                    }
                case 'S':
                    dispatch("TOGGLE-VERTICAL-MODE", gc)
                case '+':
                    dispatch("INCREASE-FONT", gc)
                case '-':
                    dispatch("REDUCE-FONT", gc)
                case 'B':
                  glim.ClearAllCaches()
                  Log2Buff("Caches cleared")
                  log.Println("Caches cleared")

            }
        }
    }

	}
    if gc.ActiveBuffer.Formatter.Cursor > gc.ActiveBuffer.Formatter.LastDrawnCharPos {
        log.Println("Advancing screen to cursor")
        //gc.ActiveBuffer.Formatter.FirstDrawnCharPos = scanToNextLine (gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.FirstDrawnCharPos)
        //gc.ActiveBuffer.Formatter.FirstDrawnCharPos = scanToPrevLine (gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
    }

if gc.ActiveBuffer.Formatter.Cursor < 0 {
    gc.ActiveBuffer.Formatter.Cursor = 0
}
    if (gc.ActiveBuffer.Formatter.Cursor <gc.ActiveBuffer.Formatter.FirstDrawnCharPos || gc.ActiveBuffer.Formatter.Cursor > gc.ActiveBuffer.Formatter.LastDrawnCharPos) {
        scrollToCursor(gc.ActiveBuffer);
    }
}
