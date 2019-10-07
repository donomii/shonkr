// +build !windows

package main

/*
#include <stdlib.h>
int launchShell() {
int masterFd;
char* args[] = {"/bin/bash", "-i", NULL };
int procId = forkpty(&masterFd, NULL, NULL,  NULL);
if( procId == 0 ){
  execve( args[0], args, NULL);
}
return masterFd;
}
*/
import "C"

import "github.com/donomii/goof"

func startShell() (chan []byte, chan []byte) {
	fileHandle := uintptr(C.launchShell())
	shellIn, shellOut := goof.WrapHandle(fileHandle, 100)
	return shellIn, shellOut
}
