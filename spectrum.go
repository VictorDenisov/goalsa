package main

import (
	"sort"
)

func newSpectrum(buf []float64) *Spectrum {
	n := len(buf) / 2
	units := make([]SpectrumUnit, n)
	for i := 0; i < n; i++ {
		units[i].freq = i
		units[i].magn = buf[i]
	}
	return &Spectrum{units}
}

var _ sort.Interface = &Spectrum{}

type Spectrum struct {
	units []SpectrumUnit
}

func (s *Spectrum) Len() int {
	return len(s.units)
}

func (s *Spectrum) Less(i, j int) bool {
	return s.units[i].magn < s.units[j].magn
}

func (s *Spectrum) Swap(i, j int) {
	s.units[i], s.units[j] = s.units[j], s.units[i]
}

type SpectrumUnit struct {
	freq int
	magn float64
}
