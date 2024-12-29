package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

func MainLoop(fileName string) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	fileViewer := viewFile(fileName)
	defer fileViewer.Destroy()

outer:
	for {
		for event := sdl.WaitEvent(); event != nil; event = sdl.WaitEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				break outer
			default:
				fileViewer.handleEvent(event)
			}
		}
	}
}
