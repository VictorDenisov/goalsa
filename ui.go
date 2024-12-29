package main

import (
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/sdl"
)

func MainLoop(fileName string) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	defer sdl.Quit()

	fileViewer := viewFile(fileName)
	defer fileViewer.Destroy()

	done := make(chan struct{})
	renderLoopComplete := make(chan struct{})
	go RenderLoop(fileViewer, done, renderLoopComplete)
	sdl.Main(func() {
		EventLoop(fileViewer, done, renderLoopComplete)
	})
	log.Info("Waiting for completion2")
	<-renderLoopComplete
}

func EventLoop(fileViewer *FileViewer, done, complete chan struct{}) {
outer:
	for {
		for event := sdl.WaitEvent(); event != nil; event = sdl.WaitEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				done <- struct{}{}
				break outer
			default:
				fileViewer.handleEvent(event)
			}
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
