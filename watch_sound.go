package main

import (
	"fmt"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

var buffer []int16

const maxBufLen = 1500

const compressionRate = 512

func watchSound() {
	var windowSize WindowSize

	// Initialize SDL
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	window, err := sdl.CreateWindow("Realtime audio", sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
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
	clearScreen(renderer)
	renderer.Present()

	audioStream, err := OpenAudioStream("default")
	if err != nil {
		panic(err)
	}
	defer audioStream.Close()

	ticker := time.NewTicker(100 * time.Millisecond)
	eventChan := eventListener()
	//ch := filterSignalStream(compressor(audioStream.GetChan()))
	ch := compressor(filterSignalStream(audioStream.GetChan()))
	//ch := audioStream.GetChan()
outer:
	for {
		select {
		case _ = <-ticker.C:
			fmt.Printf("received ticker\n")
		receiver:
			for {
				select {
				case v := <-ch:
					fmt.Printf("%d ", v)
					buffer = append(buffer, v)
					if len(buffer) > maxBufLen {
						buffer = buffer[len(buffer)-maxBufLen : len(buffer)]
					}
				default:
					break receiver
				}
			}

			fmt.Printf("\nPopulated audio stream. Buf size %d\n", len(buffer))
			mx := arrayMaximum(buffer)

			fmt.Printf("Mx: %d\n", mx)
			zeroPosition := windowSize.Height / 2

			clearScreen(renderer)

			renderer.SetDrawColor(0, 255, 0, 255)
			for i := len(buffer) - 1; i > 0; i-- {
				x := int32((len(buffer) - 1 - i) * barWidth)
				if x+barWidth > windowSize.Width {
					break
				}
				u := int32(buffer[i])
				uh := int32(float64(windowSize.Height) / 4.0 * float64(u) / float64(mx))
				rect := &sdl.Rect{x * barWidth, zeroPosition - uh, barWidth, uh}
				renderer.FillRect(rect)
			}
			renderer.Present()

			fmt.Printf("Picture rendered\n")

		case event := <-eventChan:
			switch e := event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				break outer
			case *sdl.WindowEvent:
				if e.Event == sdl.WINDOWEVENT_RESIZED {
					windowSize.Width = e.Data1
					windowSize.Height = e.Data2
					clearScreen(renderer)
					renderer.Present()
				}
				fmt.Printf("Handling window event\n")

			}
		}

	}
}

func eventListener() chan sdl.Event {
	ch := make(chan sdl.Event)
	go func() {
		for event := sdl.WaitEvent(); event != nil; event = sdl.WaitEvent() {
			ch <- event
		}
	}()
	return ch
}

func arrayMaximum(buffer []int16) int16 {
	mx := int16(0)

	for i := 0; i < len(buffer); i++ {
		if mx < absInt16(buffer[i]) {
			mx = absInt16(buffer[i])
		}
	}

	return mx
}

func max(x, y int16) int16 {
	if x < y {
		return y
	} else {
		return x
	}
}

func clearScreen(r *sdl.Renderer) {
	r.SetDrawColor(242, 242, 242, 255)
	r.Clear()
}

func compressor(ch chan int16) (r chan int16) {
	r = make(chan int16, 20000)
	go func() {
		c := 0
		v := int16(0)
		for {
			x := <-ch
			v = max(v, x)
			c++
			if c == compressionRate {
				c = 0
				r <- v
				v = 0
			}
		}
	}()
	return r
}
