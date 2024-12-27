package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/sdl"
)

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
	lastIncompleteColumnWidth := (this.area.w - shift) % columnWidth

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
	lastI := startI + columnCount
	if lastI > int32(len(this.buf)) {
		return
	}

	for j := int32(lowerMeaningfulHarmonic); j < int32(upperMeaningfulHarmonic); j++ {
		normalizedValue := uint8(this.buf[lastI][j] / maxValue * 255)
		rect := &sdl.Rect{this.area.x + shift + (lastI-startI)*columnWidth, this.area.y + (j-lowerMeaningfulHarmonic)*cellHeight, lastIncompleteColumnWidth, cellHeight}
		renderer.SetDrawColor(255-normalizedValue, 255, 255-normalizedValue, 255)
		renderer.FillRect(rect)
	}
}
