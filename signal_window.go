package main

import (
	"math"

	log "github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/sdl"
)

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
