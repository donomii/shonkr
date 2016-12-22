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
    return x
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
        if _, err := io.ReadAtLeast(r, buf, 1); err != nil {
            //log.Fatal(err)
        }
        buffAppend(1, string(buf))
        gc.BufferList[1].Formatter.Cursor = len(gc.BufferList[1].Data.Text)
    }
}

func buffAppend ( buffId int, txt string ) {
        gc.BufferList[1].Data.Text = strings.Join([]string{gc.BufferList[1].Data.Text,txt}, "")
}


func SSHAgent() ssh.AuthMethod {
    if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
        return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
    }
    return nil
}

func startSshConn(buffId int, user, password, serverAndPort string) {
                    buffAppend(buffId, "Starting ssh connection\n")
                    config := &ssh.ClientConfig{
                        User: user,
                        Auth: []ssh.AuthMethod{
                            // Use the PublicKeys method for remote authentication.
                            SSHAgent(),
                        },
                    }

                    // Dial your ssh server.
                    buffAppend(buffId, fmt.Sprintf("Connecting to ssh server as user %v: ", user))
                    buffAppend(buffId, serverAndPort)
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
                    gc.BufferList[1].Formatter.TailBuffer = true

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
                   gc.ActiveBuffer.Formatter.Cursor = scanToEndOfLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
            case key.CodeLeftArrow:
                gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.Cursor-1
            case key.CodeRightArrow:
                gc.ActiveBuffer.Formatter.Cursor = gc.ActiveBuffer.Formatter.Cursor+1
            case key.CodeUpArrow:
                gc.ActiveBuffer.Formatter.Cursor = scanToPrevLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
            case key.CodeDownArrow:
                gc.ActiveBuffer.Formatter.Cursor = scanToNextLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
            case key.CodePageDown:
                    //page down
                    pageDown(gc.ActiveBuffer)
            case key.CodePageUp:
                    //gc.ActiveBuffer.Line = gc.ActiveBuffer.Line -24
                    //if gc.ActiveBuffer.Line < 0 { gc.ActiveBuffer.Line = 0 }
                    //Page up
                    pageUp(gc.ActiveBuffer, screenWidth, screenHeight)
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
                        gc.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s",gc.ActiveBuffer.Data.Text[:gc.ActiveBuffer.Formatter.Cursor], string(e.Rune),gc.ActiveBuffer.Data.Text[gc.ActiveBuffer.Formatter.Cursor:])
                        gc.ActiveBuffer.Formatter.Cursor++
                }

            }
        } else {
            switch e.Code {
            case key.CodeA:
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
                    gc.ActiveBufferId ++
                    if gc.ActiveBufferId>len(gc.BufferList)-1 {
                        gc.ActiveBufferId = 0
                    }
                    gc.ActiveBuffer = gc.BufferList[gc.ActiveBufferId]
                    log.Printf("Next buffer: %v", gc.ActiveBufferId)
                case 'p':
                    text, _ := clipboard.ReadAll()
                    gc.ActiveBuffer.Data.Text = fmt.Sprintf("%s%s%s",gc.ActiveBuffer.Data.Text[:gc.ActiveBuffer.Formatter.Cursor], text,gc.ActiveBuffer.Data.Text[gc.ActiveBuffer.Formatter.Cursor:])
                case 'y':
                    clipboard.WriteAll(gc.ActiveBuffer.Data.Text[gc.ActiveBuffer.Formatter.SelectStart:gc.ActiveBuffer.Formatter.SelectEnd])
                case '~':
                    saveFile(gc.ActiveBuffer.Data.FileName, gc.ActiveBuffer.Data.Text)
                case 'i':
                   gc.ActiveBuffer.InputMode = true
                case '0':
                   gc.ActiveBuffer.Formatter.Cursor = SOL(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
                case '^':
                   gc.ActiveBuffer.Formatter.Cursor = SOT(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
                case '$':
                   gc.ActiveBuffer.Formatter.Cursor = scanToEndOfLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
                case 'A':
                   gc.ActiveBuffer.Formatter.Cursor = scanToEndOfLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
                   gc.ActiveBuffer.InputMode = true
                case 'a':
                   gc.ActiveBuffer.Formatter.Cursor++
                   gc.ActiveBuffer.InputMode = true
                case 'k':
                    gc.ActiveBuffer.Formatter.Cursor = scanToPrevLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
                case 'j':
                    gc.ActiveBuffer.Formatter.Cursor = scanToNextLine(gc.ActiveBuffer.Data.Text, gc.ActiveBuffer.Formatter.Cursor)
                case 'l':
                   gc.ActiveBuffer.Formatter.Cursor++
                case 'h':
                   gc.ActiveBuffer.Formatter.Cursor--
                case 'T':
                    gc.ActiveBuffer.Formatter.TailBuffer = true
                case 'W':
                    if gc.ActiveBuffer.Formatter.Outline {
                        gc.ActiveBuffer.Formatter.Outline = false
                    } else {
                        gc.ActiveBuffer.Formatter.Outline = true
                    }
                case 'S':
                    if gc.ActiveBuffer.Formatter.Vertical {
                        gc.ActiveBuffer.Formatter.Vertical = false
                    } else {
                        gc.ActiveBuffer.Formatter.Vertical = true
                    }
                case '+':
                  gc.ActiveBuffer.Formatter.FontSize += 1
                  glim.ClearAllCaches()
                case '-':
                  gc.ActiveBuffer.Formatter.FontSize -= 1
                  glim.ClearAllCaches()
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
    if (gc.ActiveBuffer.Formatter.Cursor <gc.ActiveBuffer.Formatter.FirstDrawnCharPos || gc.ActiveBuffer.Formatter.Cursor > gc.ActiveBuffer.Formatter.LastDrawnCharPos) {
        scrollToCursor(gc.ActiveBuffer);
    }
}
