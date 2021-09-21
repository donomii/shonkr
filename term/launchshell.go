// +build !windows

package main

/*
#include <unistd.h>
#include <util.h>
int launchShell(char * path) {
	int masterFd;
	char* args[] = {path,"--login", NULL };
	int procId = forkpty(&masterFd, NULL, NULL,  NULL);
	if( procId == 0 ){
	  execve( args[0], args, NULL);
	}
	return masterFd;
}
*/
import "C"

import "github.com/donomii/goof"

func startShell(path string) (chan []byte, chan []byte) {
	fileHandle := uintptr(C.launchShell(C.CString(path)))
	shellIn, shellOut := goof.WrapHandle(fileHandle, 100)
	return shellIn, shellOut
}
