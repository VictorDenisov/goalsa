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
		this.Render()
	case *sdl.MouseMotionEvent:
		mousePos := sdl.Point{e.X, e.Y}
		log.Tracef("Mouse position: %v\n", mousePos)
		if e.State&sdl.BUTTON_LEFT > 0 {
			this.view.Shift(int(-e.XRel))
		}

	case *sdl.MouseWheelEvent:
		keyboardState := sdl.GetModState()
		mx, my, _ := sdl.GetMouseState()
		mousePos := sdl.Point{mx, my}
		if keyboardState&sdl.KMOD_LSHIFT > 0 {
			println("Shift is pressed")
		} else {

			dx := mousePos.X - this.signalWindow.area.x
			this.view.Scale(e.Y, dx)

			log.Tracef("Scale factor: %v\n", int32(this.view.scaleFactor))

		}

	case *sdl.MouseButtonEvent:
		keyboardState := sdl.GetModState()
		mx, my, _ := sdl.GetMouseState()
		mousePos := sdl.Point{mx, my}
		if e.Type == sdl.MOUSEBUTTONUP {
			if this.rightMouseButtonDown && e.Button == sdl.BUTTON_RIGHT {
				this.rightMouseButtonDown = false
			}
		} else if e.Type == sdl.MOUSEBUTTONDOWN {
			if !this.leftMouseButtonDown && e.Button == sdl.BUTTON_LEFT {
				this.leftMouseButtonDown = true
			}
			if !this.rightMouseButtonDown && e.Button == sdl.BUTTON_RIGHT {
				this.rightMouseButtonDown = true
				this.selection.SelectBlock(mousePos)

			}
			if keyboardState&sdl.KMOD_LCTRL > 0 && e.Button == sdl.BUTTON_LEFT {
				if e.Y < this.signalWindow.area.y+this.signalWindow.area.h/2 {
					this.signalWindow.Renorm(this.signalWindow.area.y + this.signalWindow.area.h/2 - my)
				}
			}
			if keyboardState&sdl.KMOD_LCTRL > 0 && e.Button == sdl.BUTTON_RIGHT {
				this.signalWindow.norm = this.signalWindow.Max()
			}
		}
	}
}

func viewFile(audioFile string) *FileViewer {

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

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		panic(err)
	}
	fileViewer := &FileViewer{
		window,
		renderer,
		WindowSize{800, 600},
		false,
		false,
		view,
		selection,
		signalWindow,
		spectraWindow}
	return fileViewer
}

func (this *FileViewer) Render() {
	this.renderer.SetDrawColor(242, 242, 242, 255)
	this.renderer.Clear()

	this.selection.Draw(this.renderer)
	this.signalWindow.Draw(this.renderer)
	this.spectraWindow.Draw(this.renderer)

	this.renderer.Present()
}

func (this *FileViewer) Destroy() {
	this.renderer.Destroy()
	this.window.Destroy()
}
