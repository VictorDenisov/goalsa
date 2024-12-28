package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/sdl"
)

type WindowSize struct {
	Width, Height int32
}

type AreaRect struct {
	// upper left corner coordinates and width and height
	x, y, w, h int32
}

const barWidth = 1

const lowerMeaningfulHarmonic = 7
const upperMeaningfulHarmonic = 31

func minInt32(a, b int32) int32 {
	if a < b {
		return a
	} else {
		return b
	}
}

func drawSound(audioFile string) {

	var windowSize WindowSize

	var leftMouseButtonDown, rightMouseButtonDown bool
	var mousePos sdl.Point
	var clickOffset sdl.Point
	var rightClickOffset sdl.Point

	_, res, _, _, spectra, _ := processFile(
		audioFile,
		nil,
		nil,
	)
	view := NewView()
	selection := NewSelection(view, AreaRect{0, 0, 0, 0}, len(res))
	signalWindow := NewSignalWindow(res, view)
	spectraWindow := &HeatMap{spectra, AreaRect{0, 0, 0, 0}, view}

	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow(audioFile, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		800, 600, sdl.WINDOW_SHOWN|sdl.WINDOW_RESIZABLE)
	if err != nil {
		panic(err)
	}
	defer window.Destroy()

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	defer renderer.Destroy()
outer:
	for {
		for event := sdl.WaitEvent(); event != nil; event = sdl.WaitEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				break outer
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_RESIZED {
					windowSize.Width = e.Data1
					windowSize.Height = e.Data2
					signalWindow.area.w = windowSize.Width
					signalWindow.area.h = windowSize.Height / 2
					spectraWindow.area.y = windowSize.Height / 2
					spectraWindow.area.w = windowSize.Width
					spectraWindow.area.h = windowSize.Height / 2
					selection.area.w = windowSize.Width
					selection.area.h = windowSize.Height
				}
				renderer.SetDrawColor(242, 242, 242, 255)
				renderer.Clear()

				selection.Draw(renderer)
				signalWindow.Draw(renderer)
				spectraWindow.Draw(renderer)
				renderer.Present()
			case *sdl.MouseMotionEvent:
				mousePos = sdl.Point{e.X, e.Y}
				if leftMouseButtonDown {
					renderer.SetDrawColor(242, 242, 242, 255)
					renderer.Clear()

					view.Shift(int(clickOffset.X - mousePos.X))
					clickOffset.X = mousePos.X

					selection.Draw(renderer)
					signalWindow.Draw(renderer)
					spectraWindow.Draw(renderer)

					renderer.Present()
				}

			case *sdl.MouseWheelEvent:
				keyboardState := sdl.GetModState()
				mx, my, _ := sdl.GetMouseState()
				mousePos = sdl.Point{mx, my}
				if keyboardState&sdl.KMOD_LSHIFT > 0 {
					println("Shift is pressed")
				} else {
					renderer.SetDrawColor(242, 242, 242, 255)
					renderer.Clear()

					dx := mousePos.X - signalWindow.area.x
					view.Scale(e.Y, dx)

					log.Tracef("Scale factor: %v\n", int32(view.scaleFactor))

					selection.Draw(renderer)
					signalWindow.Draw(renderer)
					spectraWindow.Draw(renderer)

					renderer.Present()
				}

			case *sdl.MouseButtonEvent:
				keyboardState := sdl.GetModState()
				if e.Type == sdl.MOUSEBUTTONUP {
					if leftMouseButtonDown && e.Button == sdl.BUTTON_LEFT {
						leftMouseButtonDown = false
					}
					if rightMouseButtonDown && e.Button == sdl.BUTTON_RIGHT {
						rightMouseButtonDown = false
					}
				} else if e.Type == sdl.MOUSEBUTTONDOWN {
					if !leftMouseButtonDown && e.Button == sdl.BUTTON_LEFT {
						leftMouseButtonDown = true
						clickOffset.X = mousePos.X
						clickOffset.Y = mousePos.Y
						log.Tracef("click x: %v\n", clickOffset.X)
					}
					if !rightMouseButtonDown && e.Button == sdl.BUTTON_RIGHT {
						rightMouseButtonDown = true
						rightClickOffset.X = mousePos.X
						rightClickOffset.Y = mousePos.Y
						selection.SelectBlock(mousePos)

						renderer.SetDrawColor(242, 242, 242, 255)
						renderer.Clear()
						selection.Draw(renderer)
						signalWindow.Draw(renderer)
						spectraWindow.Draw(renderer)
						renderer.Present()
					}
					if keyboardState&sdl.KMOD_LCTRL > 0 && e.Button == sdl.BUTTON_LEFT {
						if e.Y < signalWindow.area.y+signalWindow.area.h/2 {
							renderer.SetDrawColor(242, 242, 242, 255)
							renderer.Clear()
							signalWindow.Renorm(signalWindow.area.y + signalWindow.area.h/2 - clickOffset.Y)
							selection.Draw(renderer)
							signalWindow.Draw(renderer)
							spectraWindow.Draw(renderer)
							renderer.Present()
						}
					}
					if keyboardState&sdl.KMOD_LCTRL > 0 && e.Button == sdl.BUTTON_RIGHT {
						renderer.SetDrawColor(242, 242, 242, 255)
						renderer.Clear()
						signalWindow.norm = signalWindow.Max()
						selection.Draw(renderer)
						signalWindow.Draw(renderer)
						spectraWindow.Draw(renderer)
						renderer.Present()
					}
				}
			}

		}
	}
	sdl.Quit()
}
