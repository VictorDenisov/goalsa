package main

import (
	"encoding/binary"
	"math"
	"os"

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

type SignalWindow struct {
	buf  []float64
	area AreaRect
	view *View
	norm float64
}

func NewSignalWindow(res []float64, view *View) *SignalWindow {
	selectedBlocksLen := len(res) / fragmentSize
	if len(res)%fragmentSize > 0 {
		selectedBlocksLen++
	}
	return &SignalWindow{res, AreaRect{0, 0, 0, 0}, view, 0}
}

func (this *SignalWindow) Renorm(v int32) {
	this.norm = float64(v) / float64(this.area.h) * this.norm
}

func (sw *SignalWindow) Get(v int) (l, u float64) {
	l = 0
	u = 0
	start := v * sw.view.scaleFactor
	end := (v + 1) * sw.view.scaleFactor
	if end < start {
		start, end = end, start
	}
	for i := start; i < end; i++ {
		if i >= len(sw.buf) {
			break
		}
		if i < 0 {
			break
		}
		if sw.buf[i] > u {
			u = sw.buf[i]
		}
		if sw.buf[i] < l {
			l = sw.buf[i]
		}
	}
	return l, u
}

func (sw *SignalWindow) Max() (mx float64) {
	for i := 0; i < len(sw.buf); i++ {
		if math.Abs(sw.buf[i]) > mx {
			mx = math.Abs(sw.buf[i])
		}
	}
	return
}

func (this *SignalWindow) Normalize(v float64) int32 {
	return int32(float64(this.area.h) / 2.0 * float64(v) / float64(this.norm))
}

func (sw *SignalWindow) Draw(renderer *sdl.Renderer) {
	renderer.SetDrawColor(0, 255, 0, 255)
	if math.Abs(sw.norm) < 0.00001 {
		sw.norm = sw.Max()
	}
	log.Tracef("Start: %v\n", sw.view.start)
	for i := sw.view.start / sw.view.scaleFactor; i < sw.view.start/sw.view.scaleFactor+int(sw.area.w)/barWidth; i++ {
		l, u := sw.Get(i)
		lh := sw.Normalize(l)
		uh := sw.Normalize(u)
		x := int32(i - sw.view.start/sw.view.scaleFactor)
		rect := &sdl.Rect{sw.area.x + x*barWidth, sw.area.y + sw.area.h/2 - uh, barWidth, uh - lh}
		renderer.FillRect(rect)
	}
}

type HeatMap struct {
	buf  [][]float64
	area AreaRect
	view *View
}

func (this *HeatMap) Draw(renderer *sdl.Renderer) {
	dx := int32(this.view.start / this.view.scaleFactor)
	columnWidth := int32(fragmentSize) / int32(this.view.scaleFactor)
	startI := int32(0) // First window that is going to be rendered.
	shift := int32(0)  // Offset of the fisrt rendered window relative to area's left boundary.

	if dx > 0 {
		startI = dx / columnWidth
		if dx%columnWidth > 0 {
			startI++
		}
		if startI >= int32(len(this.buf)) {
			return
		}
		log.Tracef("StartI: %v\n", startI)
		if dx%columnWidth > 0 {
			shift = columnWidth - dx%columnWidth
		}
	} else {
		startI = 0
		shift = -dx
	}
	log.Tracef("ColumnWidth: %v\n", columnWidth)
	log.Tracef("area width: %v\n", this.area.w)
	log.Tracef("area height: %v\n", this.area.h)
	columnCount := (this.area.w - shift) / columnWidth

	log.Tracef("Column count: %v\n", columnCount)
	cellHeight := this.area.h / int32(upperMeaningfulHarmonic-lowerMeaningfulHarmonic)
	log.Tracef("cell height: %v\n", cellHeight)

	maxValue := this.buf[startI][0]
	log.Tracef("Len: %v\n", len(this.buf[startI]))
	firstMax := startI
	if shift > 0 && firstMax > 0 {
		firstMax--
	}
	for i := firstMax; i < minInt32(startI+columnCount, int32(len(this.buf))); i++ {
		for j := lowerMeaningfulHarmonic; j < upperMeaningfulHarmonic; j++ {
			if maxValue < this.buf[i][j] {
				maxValue = this.buf[i][j]
			}
		}
	}
	log.Tracef("Max value: %v\n", maxValue)

	// Draw first incomplete window.
	if startI > 0 {
		for j := int32(lowerMeaningfulHarmonic); j < int32(upperMeaningfulHarmonic); j++ {
			normalizedValue := uint8(this.buf[startI-1][j] / maxValue * 255)
			rect := &sdl.Rect{this.area.x, this.area.y + (j-lowerMeaningfulHarmonic)*cellHeight, shift, cellHeight}

			renderer.SetDrawColor(255-normalizedValue, 255, 255-normalizedValue, 255)
			renderer.FillRect(rect)
		}
	}
	for i := int32(startI); i < minInt32(startI+columnCount, int32(len(this.buf))); i++ {
		for j := int32(lowerMeaningfulHarmonic); j < int32(upperMeaningfulHarmonic); j++ {
			normalizedValue := uint8(this.buf[i][j] / maxValue * 255)
			rect := &sdl.Rect{this.area.x + shift + (i-startI)*columnWidth, this.area.y + (j-lowerMeaningfulHarmonic)*cellHeight, columnWidth, cellHeight}
			renderer.SetDrawColor(255-normalizedValue, 255, 255-normalizedValue, 255)
			renderer.FillRect(rect)
		}
	}
}

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

func readFileData(fileName string) ([]int16, error) {
	file, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	buf := make([]int16, 0)
	for {
		var v int16
		err := binary.Read(file, binary.LittleEndian, &v)
		if err != nil {
			break
		}
		buf = append(buf, v)
	}
	return buf, nil
}

func abs16(v int16) int16 {
	if v < 0 {
		return -v
	}
	return v
}

func shrinkSignal(buf []float64) (r []float64) {
	n := len(buf) / 2
	r = make([]float64, n)
	for i := 0; i < n; i++ {
		r[i] = (buf[2*i] + buf[2*i+1]) / 2
	}
	return
}
