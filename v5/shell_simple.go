package main

import (
	"bufio"
	"log"
	"os"
	"os/exec"
)

func startShell(shellPath string) (chan []byte, chan []byte) {
	log.Printf("Starting shell: %s", shellPath)
	
	shellIn := make(chan []byte, 100)
	shellOut := make(chan []byte, 100)
	
	// Start the shell process
	cmd := exec.Command(shellPath)
	cmd.Env = os.Environ()
	
	// Get pipes for stdin/stdout
	stdin, err := cmd.StdinPipe()
	if err != nil {
		log.Printf("Failed to get stdin pipe: %v", err)
		return shellIn, shellOut
	}
	
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Failed to get stdout pipe: %v", err)
		return shellIn, shellOut
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		log.Printf("Failed to get stderr pipe: %v", err)
		return shellIn, shellOut
	}
	
	// Start the shell
	err = cmd.Start()
	if err != nil {
		log.Printf("Failed to start shell: %v", err)
		return shellIn, shellOut
	}
	
	// Handle input to shell
	go func() {
		for data := range shellIn {
			_, err := stdin.Write(data)
			if err != nil {
				log.Printf("Error writing to shell: %v", err)
				return
			}
		}
	}()
	
	// Handle output from shell (stdout)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Bytes()
			lineCopy := make([]byte, len(line)+1)
			copy(lineCopy, line)
			lineCopy[len(line)] = '\n'
			
			select {
			case shellOut <- lineCopy:
			default:
				log.Println("Shell output buffer full, dropping data")
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading shell stdout: %v", err)
		}
	}()
	
	// Handle stderr from shell
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			line := scanner.Bytes()
			lineCopy := make([]byte, len(line)+1)
			copy(lineCopy, line)
			lineCopy[len(line)] = '\n'
			
			select {
			case shellOut <- lineCopy:
			default:
				log.Println("Shell stderr buffer full, dropping data")
			}
		}
		if err := scanner.Err(); err != nil {
			log.Printf("Error reading shell stderr: %v", err)
		}
	}()
	
	// Wait for shell to exit
	go func() {
		err := cmd.Wait()
		if err != nil {
			log.Printf("Shell exited with error: %v", err)
		} else {
			log.Println("Shell exited normally")
		}
		close(shellOut)
	}()
	
	return shellIn, shellOut
}
