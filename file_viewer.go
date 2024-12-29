package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/veandco/go-sdl2/sdl"
)

type FileViewer struct {
	window   *sdl.Window
	renderer *sdl.Renderer

	windowSize WindowSize

	leftMouseButtonDown, rightMouseButtonDown bool
	clickOffset                               sdl.Point
	rightClickOffset                          sdl.Point

	view          *View
	selection     *Selection
	signalWindow  *SignalWindow
	spectraWindow *HeatMap
}

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

func (this *FileViewer) handleEvent(event sdl.Event) {
	switch e := event.(type) {
	case *sdl.WindowEvent:
		if e.Event == sdl.WINDOWEVENT_RESIZED {
			this.windowSize.Width = e.Data1
			this.windowSize.Height = e.Data2
			this.signalWindow.area.w = this.windowSize.Width
			this.signalWindow.area.h = this.windowSize.Height / 2
			this.spectraWindow.area.y = this.windowSize.Height / 2
			this.spectraWindow.area.w = this.windowSize.Width
			this.spectraWindow.area.h = this.windowSize.Height / 2
			this.selection.area.w = this.windowSize.Width
			this.selection.area.h = this.windowSize.Height
		}
		this.renderer.SetDrawColor(242, 242, 242, 255)
		this.renderer.Clear()

		this.selection.Draw(this.renderer)
		this.signalWindow.Draw(this.renderer)
		this.spectraWindow.Draw(this.renderer)
		this.renderer.Present()
	case *sdl.MouseMotionEvent:
		mousePos := sdl.Point{e.X, e.Y}
		log.Tracef("Mouse position: %v\n", mousePos)
		if this.leftMouseButtonDown {
			this.renderer.SetDrawColor(242, 242, 242, 255)
			this.renderer.Clear()

			this.view.Shift(int(this.clickOffset.X - mousePos.X))
			this.clickOffset.X = mousePos.X

			this.selection.Draw(this.renderer)
			this.signalWindow.Draw(this.renderer)
			this.spectraWindow.Draw(this.renderer)

			this.renderer.Present()
		}

	case *sdl.MouseWheelEvent:
		keyboardState := sdl.GetModState()
		mx, my, _ := sdl.GetMouseState()
		mousePos := sdl.Point{mx, my}
		if keyboardState&sdl.KMOD_LSHIFT > 0 {
			println("Shift is pressed")
		} else {
			this.renderer.SetDrawColor(242, 242, 242, 255)
			this.renderer.Clear()

			dx := mousePos.X - this.signalWindow.area.x
			this.view.Scale(e.Y, dx)

			log.Tracef("Scale factor: %v\n", int32(this.view.scaleFactor))

			this.selection.Draw(this.renderer)
			this.signalWindow.Draw(this.renderer)
			this.spectraWindow.Draw(this.renderer)

			this.renderer.Present()
		}

	case *sdl.MouseButtonEvent:
		keyboardState := sdl.GetModState()
		mx, my, _ := sdl.GetMouseState()
		mousePos := sdl.Point{mx, my}
		if e.Type == sdl.MOUSEBUTTONUP {
			if this.leftMouseButtonDown && e.Button == sdl.BUTTON_LEFT {
				this.leftMouseButtonDown = false
			}
			if this.rightMouseButtonDown && e.Button == sdl.BUTTON_RIGHT {
				this.rightMouseButtonDown = false
			}
		} else if e.Type == sdl.MOUSEBUTTONDOWN {
			if !this.leftMouseButtonDown && e.Button == sdl.BUTTON_LEFT {
				this.leftMouseButtonDown = true
				this.clickOffset.X = mousePos.X
				this.clickOffset.Y = mousePos.Y
				log.Tracef("click x: %v\n", this.clickOffset.X)
			}
			if !this.rightMouseButtonDown && e.Button == sdl.BUTTON_RIGHT {
				this.rightMouseButtonDown = true
				this.rightClickOffset.X = mousePos.X
				this.rightClickOffset.Y = mousePos.Y
				this.selection.SelectBlock(mousePos)

				this.renderer.SetDrawColor(242, 242, 242, 255)
				this.renderer.Clear()
				this.selection.Draw(this.renderer)
				this.signalWindow.Draw(this.renderer)
				this.spectraWindow.Draw(this.renderer)
				this.renderer.Present()
			}
			if keyboardState&sdl.KMOD_LCTRL > 0 && e.Button == sdl.BUTTON_LEFT {
				if e.Y < this.signalWindow.area.y+this.signalWindow.area.h/2 {
					this.renderer.SetDrawColor(242, 242, 242, 255)
					this.renderer.Clear()
					this.signalWindow.Renorm(this.signalWindow.area.y + this.signalWindow.area.h/2 - this.clickOffset.Y)
					this.selection.Draw(this.renderer)
					this.signalWindow.Draw(this.renderer)
					this.spectraWindow.Draw(this.renderer)
					this.renderer.Present()
				}
			}
			if keyboardState&sdl.KMOD_LCTRL > 0 && e.Button == sdl.BUTTON_RIGHT {
				this.renderer.SetDrawColor(242, 242, 242, 255)
				this.renderer.Clear()
				this.signalWindow.norm = this.signalWindow.Max()
				this.selection.Draw(this.renderer)
				this.signalWindow.Draw(this.renderer)
				this.spectraWindow.Draw(this.renderer)
				this.renderer.Present()
			}
		}
	}
}

func viewFile(audioFile string) {

	_, res, _, _, spectra, _ := processFile(
		audioFile,
		nil,
		nil,
	)
	view := NewView()
	selection := NewSelection(view, AreaRect{0, 0, 0, 0}, len(res))
	signalWindow := NewSignalWindow(res, view)
	spectraWindow := &HeatMap{spectra, AreaRect{0, 0, 0, 0}, view}

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
	fileViewer := FileViewer{
		window,
		renderer,
		WindowSize{800, 600},
		false,
		false,
		sdl.Point{0, 0},
		sdl.Point{0, 0},
		view,
		selection,
		signalWindow,
		spectraWindow}
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
