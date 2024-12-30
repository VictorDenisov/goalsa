package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/sdl"
)

func MainLoop(fileName string) {
	done := make(chan struct{})
	renderLoopComplete := make(chan struct{})
	sdl.Main(func() {
		sdl.Do(func() {
			if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
				panic(err)
			}
		})
		defer sdl.Do(func() { sdl.Quit() })

		var fileViewer *FileViewer
		sdl.Do(func() {
			fileViewer = viewFile(fileName)
		})
		defer sdl.Do(func() { fileViewer.Destroy() })

		go RenderLoop(fileViewer, done, renderLoopComplete)
		EventLoop(fileViewer, done, renderLoopComplete)
	})
	log.Info("Waiting for completion2")
	<-renderLoopComplete
}

func EventLoop(fileViewer *FileViewer, done, complete chan struct{}) {
outer:
	for {
		var event sdl.Event
		sdl.Do(func() {
			event = sdl.WaitEvent()
		})
		for event != nil {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				done <- struct{}{}
				break outer
			default:
				sdl.Do(func() {
					fileViewer.handleEvent(event)
				})
			}
			sdl.Do(func() {
				event = sdl.WaitEvent()
			})
		}
	}
}

const (
	Fps = 60
)

func RenderLoop(fileViewer *FileViewer, done, complete chan struct{}) {
	ticker := time.NewTicker(1000 / Fps * time.Millisecond)
outer:
	for {
		select {
		case <-ticker.C:
			sdl.Do(func() { fileViewer.Render() })
		case <-done:
			break outer
		}
	}
	complete <- struct{}{}
}
