package main

import (
	log "github.com/sirupsen/logrus"
)

type View struct {
	start       int
	scaleFactor int
}

func NewView() *View {
	return &View{0, 1}
}

func (this *View) Scale(s int32, dx int32) {
	cursorRelativeToArea := dx / barWidth
	fixedPoint := this.start + int(cursorRelativeToArea)*this.scaleFactor
	if int(s) < 0 {
		this.scaleFactor <<= int(-s)
	} else {
		this.scaleFactor >>= int(s)
	}
	if this.scaleFactor < 1 {
		this.scaleFactor = 1
	}
	this.start = (fixedPoint/this.scaleFactor - int(cursorRelativeToArea)) * this.scaleFactor
}

func (this *View) Shift(d int) {
	log.Tracef("Shift by: %v\n", d)
	log.Tracef("Scale factor: %v\n", this.scaleFactor)
	this.start = this.start + d/barWidth*this.scaleFactor
}
