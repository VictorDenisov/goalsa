package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"

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
	buf         []float64
	area        AreaRect
	start       int
	scaleFactor int
	norm        float64
}

func (this *SignalWindow) Shift(d int) {
	fmt.Printf("Shift by: %v\n", d)
	fmt.Printf("Scale factor: %v\n", this.scaleFactor)
	this.start = this.start + d/barWidth*this.scaleFactor
	if this.start < 0 {
		this.start = 0
	}
}

func (sw *SignalWindow) Get(v int) (l, u float64) {
	l = 0
	u = 0
	start := v * sw.scaleFactor
	end := (v + 1) * sw.scaleFactor
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
	fmt.Printf("Start: %v\n", sw.start)
	for i := sw.start / sw.scaleFactor; i < sw.start/sw.scaleFactor+int(sw.area.w)/barWidth; i++ {
		l, u := sw.Get(i)
		lh := sw.Normalize(l)
		uh := sw.Normalize(u)
		x := int32(i - sw.start/sw.scaleFactor)
		rect := &sdl.Rect{sw.area.x + x*barWidth, sw.area.y + sw.area.h/2 - uh, barWidth, uh - lh}
		renderer.FillRect(rect)
	}
}

type HeatMap struct {
	buf         [][]float64
	area        AreaRect
	columnWidth int32
	dx          int32
}

func (this *HeatMap) Draw(renderer *sdl.Renderer) {
	startI := this.dx / this.columnWidth
	if this.dx%this.columnWidth > 0 {
		startI++
	}
	if startI >= int32(len(this.buf)) {
		return
	}
	fmt.Printf("StartI: %v\n", startI)
	shift := int32(0)
	if this.dx%this.columnWidth > 0 {
		shift = this.columnWidth - this.dx%this.columnWidth
	}
	fmt.Printf("ColumnWidth: %v\n", this.columnWidth)
	fmt.Printf("area width: %v\n", this.area.w)
	fmt.Printf("area height: %v\n", this.area.h)
	columnCount := (this.area.w - shift) / this.columnWidth

	fmt.Printf("Column count: %v\n", columnCount)
	cellHeight := this.area.h / int32(upperMeaningfulHarmonic-lowerMeaningfulHarmonic)
	fmt.Printf("cell height: %v\n", cellHeight)

	maxValue := this.buf[startI][0]
	fmt.Printf("Len: %v\n", len(this.buf[startI]))
	for i := startI; i < minInt32(startI+columnCount, int32(len(this.buf))); i++ {
		for j := lowerMeaningfulHarmonic; j < upperMeaningfulHarmonic; j++ {
			if maxValue < this.buf[i][j] {
				maxValue = this.buf[i][j]
			}
		}
	}
	fmt.Printf("Max value: %v\n", maxValue)

	for i := int32(startI); i < minInt32(startI+columnCount, int32(len(this.buf))); i++ {
		for j := int32(lowerMeaningfulHarmonic); j < int32(upperMeaningfulHarmonic); j++ {
			normalizedValue := uint8(this.buf[i][j] / maxValue * 255)
			rect := &sdl.Rect{this.area.x + shift + (i-startI)*this.columnWidth, this.area.y + (j-lowerMeaningfulHarmonic)*cellHeight, this.columnWidth, cellHeight}
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
	view := &SignalWindow{res, AreaRect{0, 0, 0, 0}, 0, 1, 0}
	spectraWindow := &HeatMap{spectra, AreaRect{0, 0, 0, 0}, fragmentSize, 0}
	selectedBlocksLen := len(res) / fragmentSize
	if len(res)%fragmentSize > 0 {
		selectedBlocksLen++
	}
	selectedBlocks := make([]bool, selectedBlocksLen)
	/*
		buf, err := readFileData(audioFile)
		if err != nil {
			panic(err)
		}
	*/

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
					view.area.w = windowSize.Width
					view.area.h = windowSize.Height / 2
					spectraWindow.area.y = windowSize.Height / 2
					spectraWindow.area.w = windowSize.Width
					spectraWindow.area.h = windowSize.Height / 2

				}
				renderer.SetDrawColor(242, 242, 242, 255)
				renderer.Clear()

				view.Draw(renderer)
				spectraWindow.Draw(renderer)
				renderer.Present()
			case *sdl.MouseMotionEvent:
				mousePos = sdl.Point{e.X, e.Y}
				keyboardState := sdl.GetModState()
				if leftMouseButtonDown {
					if keyboardState&sdl.KMOD_LCTRL > 0 {
					} else {
						renderer.SetDrawColor(242, 242, 242, 255)
						renderer.Clear()

						view.Shift(int(clickOffset.X - mousePos.X))
						clickOffset.X = mousePos.X
						view.Draw(renderer)

						spectraWindow.dx = int32(view.start / view.scaleFactor)
						if spectraWindow.dx < 0 {
							spectraWindow.dx = 0
						}
						spectraWindow.Draw(renderer)
						renderer.Present()
					}
				}

			case *sdl.MouseWheelEvent:
				keyboardState := sdl.GetModState()
				if keyboardState&sdl.KMOD_LSHIFT > 0 {
					println("Shift is pressed")
				} else {
					renderer.SetDrawColor(242, 242, 242, 255)
					renderer.Clear()

					if int(e.Y) < 0 {
						view.scaleFactor <<= int(-e.Y)
					} else {
						view.scaleFactor >>= int(e.Y)
					}
					if view.scaleFactor < 1 {
						view.scaleFactor = 1
					}
					view.Draw(renderer)

					spectraWindow.columnWidth = int32(fragmentSize) / int32(view.scaleFactor)

					fmt.Printf("Scale factor: %v\n", int32(view.scaleFactor))

					spectraWindow.Draw(renderer)
					renderer.Present()
				}

			case *sdl.MouseButtonEvent:
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
						fmt.Printf("click x: %v\n", clickOffset.X)
					}
					if !rightMouseButtonDown && e.Button == sdl.BUTTON_RIGHT {
						rightMouseButtonDown = true
						rightClickOffset.X = mousePos.X
						rightClickOffset.Y = mousePos.Y
						selectedBlock := (int(rightClickOffset.X)*view.scaleFactor + view.start) / fragmentSize
						selectedBlocks[selectedBlock] = true
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
