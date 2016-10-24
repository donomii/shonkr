// +build !VR

package main

import (
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

func scanToEndOfLine (txt string, c int) int{
    letters := strings.Split(txt, "")
    x:=c
    for x= c+1; x>0 && x<len(txt) && !( letters[x]== "\n"); x++ {}
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

func processPort(r io.Reader) {
    for {
        buf := make([]byte, 1)
        if _, err := io.ReadAtLeast(r, buf, 1); err != nil {
            //log.Fatal(err)
        }
        gc.ActiveBuffer.Text = strings.Join([]string{gc.ActiveBuffer.Text,string(buf)}, "")
    }
}


func SSHAgent() ssh.AuthMethod {
    if sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK")); err == nil {
        return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers)
    }
    return nil
}

func handleEvent(a app.App, i interface{}) {
	log.Println(i)
	switch e := a.Filter(i).(type) {
	case key.Event:
     switch e.Code {
            case key.CodeDeleteBackspace:
                if gc.ActiveBuffer.Cursor > 0 {
                    gc.ActiveBuffer.Text = deleteLeft(gc.ActiveBuffer.Text, gc.ActiveBuffer.Cursor)
                    gc.ActiveBuffer.Cursor--
                }
            case key.CodeLeftArrow:
                gc.ActiveBuffer.Cursor = gc.ActiveBuffer.Cursor-1
            case key.CodeRightArrow:
                gc.ActiveBuffer.Cursor = gc.ActiveBuffer.Cursor+1
            case key.CodeUpArrow:
                gc.ActiveBuffer.Cursor = scanToPrevLine(gc.ActiveBuffer.Text, gc.ActiveBuffer.Cursor)
            case key.CodeDownArrow:
                gc.ActiveBuffer.Cursor = scanToNextLine(gc.ActiveBuffer.Text, gc.ActiveBuffer.Cursor)
            case key.CodePageDown:
                    //page down
                    log.Println("Scanning to start of next page from ", activeFormatter.LastDrawnCharPos)
                    activeFormatter.FirstDrawnCharPos = scanToPrevLine(gc.ActiveBuffer.Text,activeFormatter.LastDrawnCharPos)
                    gc.ActiveBuffer.Cursor = activeFormatter.FirstDrawnCharPos
            case key.CodePageUp:
                    //gc.ActiveBuffer.Line = gc.ActiveBuffer.Line -24
                    //if gc.ActiveBuffer.Line < 0 { gc.ActiveBuffer.Line = 0 }
                    //Page up
                    start := searchBackPage(gc.ActiveBuffer.Text, activeFormatter)
                    log.Println("New start at ", start)
                    activeFormatter.FirstDrawnCharPos = start
                    gc.ActiveBuffer.Cursor = activeFormatter.FirstDrawnCharPos
   }
        if gc.ActiveBuffer.InputMode {
            switch e.Code {
            case key.CodeLeftShift:
            case key.CodeRightShift:
            case key.CodeReturnEnter:
                gc.ActiveBuffer.Text = fmt.Sprintf("%s%s%s",gc.ActiveBuffer.Text[:gc.ActiveBuffer.Cursor], "\n",gc.ActiveBuffer.Text[gc.ActiveBuffer.Cursor:])
                gc.ActiveBuffer.Cursor++
            default:
                switch e.Rune {
                    case '`':
                        gc.ActiveBuffer.InputMode = false
                    default:
                        gc.ActiveBuffer.Text = fmt.Sprintf("%s%s%s",gc.ActiveBuffer.Text[:gc.ActiveBuffer.Cursor], string(e.Rune),gc.ActiveBuffer.Text[gc.ActiveBuffer.Cursor:])
                        gc.ActiveBuffer.Cursor++
                }

            }
        } else {
            switch e.Code {
            case key.CodeA:
                gc.ActiveBuffer.Cursor = gc.ActiveBuffer.Cursor-1
            case key.CodeD:
                gc.ActiveBuffer.Cursor = gc.ActiveBuffer.Cursor+1
            case key.CodeQ:
                gc.ActiveBuffer.Cursor = gc.ActiveBuffer.Cursor +1
            case key.CodeE:
                gc.ActiveBuffer.Cursor = gc.ActiveBuffer.Cursor -1
            }
            switch e.Rune {
                case 'L':
                    config := &ssh.ClientConfig{
                        User: "",
                        Auth: []ssh.AuthMethod{
                            // Use the PublicKeys method for remote authentication.
                            SSHAgent(),
                        },
                    }

                    // Dial your ssh server.
                    conn, err := ssh.Dial("tcp", ":22", config)
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

                    stderr, err := session.StderrPipe()
                    if err != nil {
                        //return fmt.Errorf("Unable to setup stderr for session: %v", err)
                    }
                    go io.Copy(os.Stderr, stderr)

                    err = session.Run("ls -l $LC_USR_DIR")
                    defer conn.Close()

                case 'n':
                    gc.ActiveBufferId ++
                    if gc.ActiveBufferId>len(gc.BufferList)-1 {
                        gc.ActiveBufferId = 0
                    }
                    gc.ActiveBuffer = gc.BufferList[gc.ActiveBufferId]
                    log.Printf("Next buffer: %v", gc.ActiveBufferId)
                case 'v':
                    text, _ := clipboard.ReadAll()
                        gc.ActiveBuffer.Text = fmt.Sprintf("%s%s%s",gc.ActiveBuffer.Text[:gc.ActiveBuffer.Cursor], text,gc.ActiveBuffer.Text[gc.ActiveBuffer.Cursor:])
                case '~':
                    saveFile(fname, gc.ActiveBuffer.Text)
                case 'i':
                   gc.ActiveBuffer.InputMode = true
                case '$':
                   gc.ActiveBuffer.Cursor = scanToEndOfLine(gc.ActiveBuffer.Text, gc.ActiveBuffer.Cursor)
                case 'A':
                   gc.ActiveBuffer.Cursor = scanToEndOfLine(gc.ActiveBuffer.Text, gc.ActiveBuffer.Cursor)
                   gc.ActiveBuffer.InputMode = true
                case 'a':
                   gc.ActiveBuffer.Cursor++
                   gc.ActiveBuffer.InputMode = true
                case 'w':
                    gc.ActiveBuffer.Cursor = scanToPrevLine(gc.ActiveBuffer.Text, gc.ActiveBuffer.Cursor)
                case 's':
                    gc.ActiveBuffer.Cursor = scanToNextLine(gc.ActiveBuffer.Text, gc.ActiveBuffer.Cursor)
                case 'W':
                case 'S':
            }
        }

	}
    if gc.ActiveBuffer.Cursor > activeFormatter.LastDrawnCharPos {
        activeFormatter.FirstDrawnCharPos = scanToNextLine (gc.ActiveBuffer.Text, activeFormatter.FirstDrawnCharPos)
    }
}

