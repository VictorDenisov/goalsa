package main

import (
	"encoding/binary"
	"math"
	"os"

	"github.com/veandco/go-sdl2/sdl"
)

type WindowSize struct {
	Width, Height int32
}

type SignalWindow struct {
	buf         []float64
	start       int
	scaleFactor int
}

func (sw *SignalWindow) Get(v int) (l, u float64) {
	l = 0
	u = 0
	start := sw.start + v*sw.scaleFactor
	end := sw.start + (v+1)*sw.scaleFactor
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

const barWidth = 1

func (sw *SignalWindow) Draw(zeroPosition int32, renderer *sdl.Renderer, windowSize WindowSize) {
	renderer.SetDrawColor(0, 255, 0, 255)
	mx := sw.Max()
	lb := -int(windowSize.Width) / (4 * barWidth)
	for i := lb; i < int(windowSize.Width)/(4*barWidth); i++ {
		l, u := sw.Get(i)
		lh := int32(float64(windowSize.Height) / 4.0 * float64(l) / float64(mx))
		uh := int32(float64(windowSize.Height) / 4.0 * float64(u) / float64(mx))
		x := int32(i - lb)
		rect := &sdl.Rect{x * barWidth * 2, zeroPosition - uh, barWidth, uh - lh}
		renderer.FillRect(rect)
	}
	renderer.Present()
}

func drawSound(audioFile string) {

	var windowSize WindowSize

	var leftMouseButtonDown, rightMouseButtonDown bool
	var mousePos sdl.Point
	var clickOffset sdl.Point
	var rightClickOffset sdl.Point

	_, res, _, spectra, _ := processFile(
		audioFile,
		nil,
		nil,
	)
	view := &SignalWindow{res, 0, 1}
	spectraWindow := &SignalWindow{spectra, 0, 1}
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
	var lastOffset int

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

				}
				renderer.SetDrawColor(242, 242, 242, 255)
				renderer.Clear()

				view.Draw(windowSize.Height/4, renderer, windowSize)
				spectraWindow.Draw(windowSize.Height/4*3, renderer, windowSize)
			case *sdl.MouseMotionEvent:
				mousePos = sdl.Point{e.X, e.Y}
				if leftMouseButtonDown {

					renderer.SetDrawColor(242, 242, 242, 255)
					renderer.Clear()

					view.start = lastOffset - int(mousePos.X-clickOffset.X)/barWidth/2*view.scaleFactor
					if view.start < 0 {
						view.start = 0
					}
					view.Draw(windowSize.Height/4, renderer, windowSize)

					spectraWindow.start = lastOffset - int(mousePos.X-clickOffset.X)/barWidth/2*spectraWindow.scaleFactor
					if spectraWindow.start < 0 {
						spectraWindow.start = 0
					}
					spectraWindow.Draw(windowSize.Height/4*3, renderer, windowSize)
				}

			case *sdl.MouseWheelEvent:
				renderer.SetDrawColor(242, 242, 242, 255)
				renderer.Clear()

				view.scaleFactor -= int(e.Y)
				if view.scaleFactor < 1 {
					view.scaleFactor = 1
				}
				view.Draw(windowSize.Height/4, renderer, windowSize)

				spectraWindow.scaleFactor -= int(e.Y)
				if spectraWindow.scaleFactor < 1 {
					spectraWindow.scaleFactor = 1
				}
				spectraWindow.Draw(windowSize.Height/4*3, renderer, windowSize)
			case *sdl.MouseButtonEvent:
				if e.Type == sdl.MOUSEBUTTONUP {
					if leftMouseButtonDown && e.Button == sdl.BUTTON_LEFT {
						leftMouseButtonDown = false
						lastOffset = lastOffset - int(mousePos.X-clickOffset.X)/barWidth/2*view.scaleFactor
					}
					if rightMouseButtonDown && e.Button == sdl.BUTTON_RIGHT {
						rightMouseButtonDown = false
					}
				} else if e.Type == sdl.MOUSEBUTTONDOWN {
					if !leftMouseButtonDown && e.Button == sdl.BUTTON_LEFT {
						leftMouseButtonDown = true
						clickOffset.X = mousePos.X
						clickOffset.Y = mousePos.Y
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
