package main

import (
	"log"
	"os"
	"os/exec"
)

// startShellWithBackend starts a shell and connects it to the terminal backend
func startShellWithBackend(shellPath string, term *TerminalBackend) (*os.File, error) {
	cmd := exec.Command(shellPath)
	cmd.Env = os.Environ()

	width, height := term.GetSize()

	// Start PTY
	ptymaster, err := StartPty(cmd, width, height)
	if err != nil {
		return nil, err
	}

	// Copy PTY output to terminal
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := ptymaster.Read(buf)
			if err != nil {
				log.Printf("PTY read error: %v", err)
				return
			}
			if n > 0 {
				term.Write(buf[:n])
				needsRedraw = true
			}
		}
	}()

	// Handle shell input -> PTY input
	go func() {
		for data := range shellIn {
			_, err := ptymaster.Write(data)
			if err != nil {
				log.Printf("PTY write error: %v", err)
				return
			}
		}
	}()

	// Wait for shell to exit
	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Printf("Shell exited with error: %v", err)
		}
		ptymaster.Close()
	}()

	return ptymaster, nil
}
