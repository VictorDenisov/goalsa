package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

func MainLoop(fileName string) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	viewFile(fileName)
}
