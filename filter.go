package main

import (
	"github.com/mjibson/go-dsp/dsputils"
	"github.com/mjibson/go-dsp/fft"
)

var filter []complex128

func init() {
	filter = dsputils.ZeroPad([]complex128{5}, 5)
}

type Filter struct {
	kernel    []float64
	fft       []complex128
	blockSize int
	rem       []float64
}

func NewHpFilter(m int, fc float64, blockSize int) *Filter {
	kernel := dsputils.ZeroPadF(windowSincKernelHp(m, fc), m+blockSize)
	return &Filter{kernel, fft.FFTReal(kernel), blockSize, []float64{}}
}

func NewLpFilter(m int, fc float64, blockSize int) *Filter {
	kernel := dsputils.ZeroPadF(windowSincKernelLp(m, fc), m+blockSize)
	return &Filter{kernel, fft.FFTReal(kernel), blockSize, []float64{}}
}

func NewBpFilter(m int, fcL float64, fcH float64, blockSize int) *Filter {
	kernel := dsputils.ZeroPadF(windowSincKernelBp(m, fcL, fcH), m+blockSize)
	return &Filter{kernel, fft.FFTReal(kernel), blockSize, []float64{}}
}

func (f *Filter) Convolve(signal []float64) []float64 {
	signal = dsputils.ZeroPadF(signal, len(f.fft))
	fft_y := fft.FFTReal(signal)

	r := make([]complex128, len(signal))
	for i := 0; i < len(r); i++ {
		r[i] = f.fft[i] * fft_y[i]
	}

	return ToReal(fft.IFFT(r))
}

func (f *Filter) FilterBuf(buf []float64) []float64 {
	res := f.Convolve(buf)
	for i := 0; i < len(f.rem); i++ {
		res[i] += f.rem[i]
	}
	sig := res[0:f.blockSize]
	f.rem = res[f.blockSize:len(res)]

	return sig
}
