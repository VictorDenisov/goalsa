package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

func drawSound(audioFile string) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("test", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		800, 600, sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, 0)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()
	running := true
	for running {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			sdl.Delay(10)
			switch e := event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				running = false
				break
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_RESIZED {
					width := e.Data1
					height := e.Data2
					fmt.Printf("Current window size: %d, %d\n", width, height)
				}
			}
			renderer.SetDrawColor(242, 242, 242, 255)
			renderer.Clear()
			renderer.Present()

		}
	}
	sdl.Quit()
}
