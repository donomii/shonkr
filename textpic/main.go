package main

import (
	"flag"
	"io/ioutil"

	"github.com/donomii/glim"
)

func main() {

	var outFile string
	var message string
	var messageFile string
	var width int
	var height int
	flag.StringVar(&outFile, "file", "image.png", "Save the picture to this file")
	flag.StringVar(&message, "message", "Hello world!", "Message to render")
	flag.StringVar(&messageFile, "message-file", "", "Read message from file")
	flag.IntVar(&width, "width", 640, "Picture width")
	flag.IntVar(&height, "height", 480, "Picture height")
	flag.Parse()

	if messageFile != "" {
		contents, _ := ioutil.ReadFile(messageFile)
		message = string(contents)
	}
	p1 := make([]uint8, width*height*4)
	formatter := glim.NewFormatter()
	glim.RenderPara(formatter, 0, 0, 0, 0, width, height, width, height, 0, 0, p1, message, true, true, true)
	for i := 3; i < len(p1); i = i + 4 {
		if p1[i] > 0 {
			p1[i-3] = 0
		} else {
			p1[i-3] = 255
			p1[i-2] = 255
			p1[i-1] = 255
			p1[i] = 255
		}
	}
	p2 := glim.FlipUp(width, height, p1)

	glim.SaveImage(glim.ImageToGFormatRGBA(width, height, p2), outFile)
}
