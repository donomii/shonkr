// +build !windows

package main

/*
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

func startShell() (chan []byte, chan []byte) {
	fileHandle := uintptr(C.launchShell())
	shellIn, shellOut := goof.WrapHandle(fileHandle, 100)
	return shellIn, shellOut
}
