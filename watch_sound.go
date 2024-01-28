package main

import (
	"fmt"
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

var buffer []int16

func watchSound() {
	var windowSize WindowSize

	// Initialize SDL
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
	renderer.SetDrawColor(242, 242, 242, 255)
	renderer.Clear()
	renderer.Present()

	renderer.SetDrawColor(0, 255, 0, 255)
	rect := &sdl.Rect{0, 0, 80, 200}
	renderer.FillRect(rect)
	renderer.Present()

	audioStream, err := OpenAudioStream("default")
	if err != nil {
		panic(err)
	}
	defer audioStream.Close()

	ticker := time.NewTicker(1 * time.Second)
	eventChan := eventListener()
outer:
	for {
		select {
		case _ = <-ticker.C:
			fmt.Printf("received ticker\n")
			ch := audioStream.GetChan()
		receiver:
			for {
				select {
				case v := <-ch:
					fmt.Printf("%d ", v)
					buffer = append(buffer, v)
					if len(buffer) > 1000 {
						buffer = buffer[len(buffer)-1000 : len(buffer)]
					}
				default:
					break receiver
				}
			}

			fmt.Printf("\nPopulated audio stream. Buf size %d\n", len(buffer))
			mx := int16(0)

			for i := 0; i < len(buffer); i++ {
				if mx < absInt16(buffer[i]) {
					mx = absInt16(buffer[i])
				}
			}
			fmt.Printf("Mx: %d\n", mx)
			zeroPosition := windowSize.Height / 2
			renderer.SetDrawColor(242, 242, 242, 255)
			renderer.Clear()
			renderer.Present()
			renderer.SetDrawColor(0, 255, 0, 255)
			for i := len(buffer) - 1; i > 0; i-- {
				x := int32((len(buffer) - 1 - i) * barWidth * 2)
				if x+barWidth > windowSize.Width {
					break
				}
				u := int32(buffer[i])
				uh := int32(float64(windowSize.Height) / 4.0 * float64(u) / float64(mx))
				rect := &sdl.Rect{x * barWidth * 2, zeroPosition - uh, barWidth, uh}
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
				}
				fmt.Printf("Handling window event\n")

				renderer.SetDrawColor(242, 242, 242, 255)
				renderer.Clear()
				renderer.Present()
				renderer.SetDrawColor(0, 255, 0, 255)
				rect := &sdl.Rect{0, 0, 80, 80}
				renderer.FillRect(rect)
				renderer.Present()
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
