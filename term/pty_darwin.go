package main

/*
#include <util.h>
#include <sys/ioctl.h>
#include <errno.h>

int my_openpty(int *amaster, int *aslave) {
    return openpty(amaster, aslave, NULL, NULL, NULL);
}

int my_resize(int fd, int rows, int cols) {
    struct winsize ws;
    ws.ws_row = rows;
    ws.ws_col = cols;
    ws.ws_xpixel = 0;
    ws.ws_ypixel = 0;
    return ioctl(fd, TIOCSWINSZ, &ws);
}
*/
import "C"

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"syscall"
)

// StartPty starts a command in a pseudo-terminal
func StartPty(c *exec.Cmd, width, height int) (*os.File, error) {
	master, slave, err := openPty()
	if err != nil {
		return nil, err
	}

	c.Stdout = slave
	c.Stderr = slave
	c.Stdin = slave
	c.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
		Ctty:    0, // Child's FD 0 (Stdin) which is the slave
	}

	if err := c.Start(); err != nil {
		master.Close()
		slave.Close()
		return nil, err
	}

	log.Printf("Shell started using PTY")

	// Close slave in parent - child has its own copy
	slave.Close()

	// Resize initially
	resizePty(master, width, height)

	return master, nil
}

func openPty() (*os.File, *os.File, error) {
	var master, slave C.int
	ret := C.my_openpty(&master, &slave)
	if ret != 0 {
		return nil, nil, fmt.Errorf("openpty failed with ret=%d", ret)
	}
	return os.NewFile(uintptr(master), "pty-master"), os.NewFile(uintptr(slave), "pty-slave"), nil
}

func resizePty(f *os.File, width, height int) error {
	ret := C.my_resize(C.int(f.Fd()), C.int(height), C.int(width))
	if ret != 0 {
		return fmt.Errorf("resize ioctl failed")
	}
	return nil
}
