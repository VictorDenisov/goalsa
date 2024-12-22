package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/sdl"
)

type Selection struct {
	view           *View
	area           AreaRect
	selectedBlocks []bool
}

func NewSelection(view *View, area AreaRect, signalLen int) *Selection {
	blockCount := signalLen / fragmentSize
	if signalLen%fragmentSize > 0 {
		blockCount++
	}
	return &Selection{view, area, make([]bool, blockCount)}
}

func (this *Selection) SelectBlock(p sdl.Point) {
	globalScaledPosition := this.view.start + int(p.X)*this.view.scaleFactor
	fragmentNumber := globalScaledPosition / fragmentSize
	if globalScaledPosition%fragmentSize == 0 {
		fragmentNumber--
	}
	this.selectedBlocks[fragmentNumber] = !this.selectedBlocks[fragmentNumber]
}

func (this *Selection) Draw(renderer *sdl.Renderer) {
	dx := int32(this.view.start / this.view.scaleFactor)
	columnWidth := int32(fragmentSize) / int32(this.view.scaleFactor)
	startI := int32(0) // First window that is going to be rendered.
	shift := int32(0)  // Offset of the fisrt rendered window relative to area's left boundary.

	if dx > 0 {
		startI = dx / columnWidth
		if dx%columnWidth > 0 {
			startI++
		}
		if startI >= int32(len(this.selectedBlocks)) {
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
	columnCount := (this.area.w - shift) / columnWidth

	if startI > 0 && this.selectedBlocks[startI] {
		rect := &sdl.Rect{this.area.x, this.area.y, shift, this.area.h}

		renderer.SetDrawColor(202, 233, 245, 255)
		renderer.FillRect(rect)
	}
	for i := int32(startI); i < minInt32(startI+columnCount, int32(len(this.selectedBlocks))); i++ {
		if this.selectedBlocks[i] {
			rect := &sdl.Rect{this.area.x + shift + (i-startI)*columnWidth, this.area.y, columnWidth, this.area.h}
			renderer.SetDrawColor(202, 233, 245, 255)
			renderer.FillRect(rect)
		}
	}
}
