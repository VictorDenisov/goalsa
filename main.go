package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/cmplx"
	"os"
	"sort"
	"unsafe"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/mjibson/go-dsp/dsputils"
	"github.com/mjibson/go-dsp/fft"
	log "github.com/sirupsen/logrus"
)

/*
#cgo LDFLAGS: -lasound
#include <alsa/asoundlib.h>
#include <stdint.h>
*/
import "C"

var filter []complex128

func init() {
	filter = dsputils.ZeroPad([]complex128{5}, 5)
}

// This code captures standard input
// Redirect this files standard output into a file.
// Play the file using the following command:
// aplay -t raw -f S16_LE -c1 -r44100 $1
// 1 channel
// little endian
// 16 bit
// 44100 sampling rate
//
// Import raw data using audacity with the specified parameters
func main() {
	_, _, values, _ := processFile("short.wav")
	es := measureIntervals(values)
	s := detectCode(es)
	fmt.Printf("String: %s\n", s)

	return

	var handle *C.snd_pcm_t
	var rc C.int

	deviceCString := C.CString("default")
	defer C.free(unsafe.Pointer(deviceCString))

	rc = C.snd_pcm_open(&handle, deviceCString, C.SND_PCM_STREAM_CAPTURE, 0)
	if rc < 0 {
		log.Fatal("Unable to open pcm device: %v", C.snd_strerror(rc))
	}
	defer C.snd_pcm_close(handle)
	defer C.snd_pcm_drain(handle)

	// Setup params
	var params *C.snd_pcm_hw_params_t
	rc = C.snd_pcm_hw_params_malloc(&params)
	if rc < 0 {
		log.Fatal("Couldn't alloc hw params")
	}
	defer C.snd_pcm_hw_params_free(params)
	rc = C.snd_pcm_hw_params_any(handle, params)
	if rc < 0 {
		log.Fatal("Couldn't set default hw params")
	}
	rc = C.snd_pcm_hw_params_set_access(handle, params, C.SND_PCM_ACCESS_RW_INTERLEAVED)
	if rc < 0 {
		log.Fatal("Couldn't set access params")
	}
	rc = C.snd_pcm_hw_params_set_format(handle, params, C.SND_PCM_FORMAT_S16_LE)
	if rc < 0 {
		log.Fatal("Couldn't set endian format")
	}
	rc = C.snd_pcm_hw_params_set_channels(handle, params, 1)
	if rc < 0 {
		log.Fatal("Couldn't set channels")
	}
	var val C.uint = 44100
	var dir C.int
	rc = C.snd_pcm_hw_params_set_rate_near(handle, params, &val, &dir)
	if rc < 0 {
		log.Fatal("Couldn't set rate")
	}

	var frames C.snd_pcm_uframes_t = 32
	rc = C.snd_pcm_hw_params_set_period_size_near(handle, params, &frames, &dir)
	if rc < 0 {
		log.Fatal("Couldn't set period size")
	}

	rc = C.snd_pcm_hw_params(handle, params)
	if rc < 0 {
		log.Fatal("Couldn't set params")
	}

	rc = C.snd_pcm_hw_params_get_period_size(params, &frames, &dir)
	if rc < 0 {
		log.Fatal("Couldn't get period size")
	}

	var size C.snd_pcm_uframes_t
	log.Tracef("Frames: %v\n", frames)
	size = frames * 2
	buffer := make([]byte, size)
	log.Tracef("Buffer len: %v", len(buffer))

	rc = C.snd_pcm_hw_params_get_period_time(params, &val, &dir)
	if rc < 0 {
		log.Fatal("Couldn't get period time")
	}
	var loops C.long = 10_000_000 / C.long(val)

	out, err := os.Create("output.raw")
	if err != nil {
		log.Fatalf("Couldn't create output file: %v", err)
	}
	defer out.Close()
	for loops > 0 {
		loops--
		var rcl C.long
		rcl = C.snd_pcm_readi(handle, unsafe.Pointer(&buffer[0]), frames)
		if rcl == -C.EPIPE {
			fmt.Printf("Overrun occurred\n")
			C.snd_pcm_prepare(handle)
		} else if rcl < 0 {
			fmt.Printf("Error from read: %v\n", C.snd_strerror(rc))
		} else if rcl != C.long(frames) {
			fmt.Printf("Short read, read %v frames\n", rc)
		}
		log.Tracef("rcl read: %v\n", rcl)
		n, err := out.Write(buffer)
		if err != nil && n < len(buffer) {
			fmt.Printf("Failed to write: %v\n", err)
		}
	}

}

func ToAbs(a []complex128) []float64 {
	r := make([]float64, len(a))
	for i := 0; i < len(a); i++ {
		r[i] = cmplx.Abs(a[i])
	}
	return r
}

func ToReal(a []complex128) []float64 {
	r := make([]float64, len(a))
	for i := 0; i < len(a); i++ {
		if math.Abs(imag(a[i])) > 0.000001 {
			panic("Converting complex number with non zero imaginary part")
		}
		r[i] = real(a[i])
	}
	return r
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

func readFile(name string) (res []float64, err error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	buf := make([]float64, 88200)
	for i := 0; i < len(buf); i++ {
		var v int16
		err := binary.Read(file, binary.LittleEndian, &v)
		if err != nil {
			return res, nil
		}
		buf[i] = float64(v)
	}
	return buf, nil
}

func hann(y []float64) {
	n := len(y) - 1
	for x := 0; x < len(y); x++ {
		v := (1 - math.Cos(2*math.Pi*float64(x)/float64(n))) / 2
		y[x] *= v
	}
}

func processFile(name string) (sig []float64, res []float64, values []bool, err error) {

	values = make([]bool, 0)

	filter := NewBpFilter(200, 7.0/441, 30.0/441, 441)

	file, err := os.Open(name)
	if err != nil {
		return nil, nil, nil, err
	}
	defer file.Close()
	sig = make([]float64, 0)
	res = make([]float64, 0)
	for pieceNum := 0; ; pieceNum++ {
		buf := make([]float64, 441)
		for i := 0; i < 441; i++ {
			var v int16
			err := binary.Read(file, binary.LittleEndian, &v)
			if err != nil {
				return sig, res, values, nil
			}
			buf[i] = float64(v)
		}
		sig = append(sig, buf...)
		buf = filter.FilterBuf(buf)
		res = append(res, buf...)
		hann(buf)
		rawSpectrum := ToAbs(fft.FFTReal(buf))

		spectrum := newSpectrum(rawSpectrum)
		sort.Sort(sort.Reverse(spectrum))
		if spectrum.units[0].magn-spectrum.units[1].magn > 50000 {
			values = append(values, true)
		} else {
			values = append(values, false)
		}

	}
	return sig, res, values, nil
}

func rng(n int) []int {
	r := make([]int, n)
	for i := 0; i < n; i++ {
		r[i] = i
	}
	return r
}

func drawChart(name string, buf []float64) {
	// create a new bar instance
	bar := charts.NewBar()

	// Set global options
	bar.SetGlobalOptions(charts.WithTitleOpts(opts.Title{
		Title:    "Bar chart in Go",
		Subtitle: "This is fun to use!",
	}))
	barData := make([]opts.BarData, len(buf))
	for i := 0; i < len(buf); i++ {
		barData[i] = opts.BarData{Value: buf[i]}
	}
	bar.SetXAxis(rng(len(buf))).AddSeries("A", barData)

	// Put data into instance
	f, _ := os.Create(name)
	_ = bar.Render(f)
}

func windowSincKernelLp(m int, fc float64) []float64 {
	h := make([]float64, m+1)
	for i := 0; i <= m; i++ {
		// Blackman window
		iF := float64(i)
		mF := float64(m)
		mF2 := float64(m / 2)
		w := 0.42 - 0.5*math.Cos(2*math.Pi*iF/mF) + 0.08*math.Cos(4*math.Pi*iF/mF)
		h[i] = w * math.Sin(2*math.Pi*fc*(iF-mF2)) / (iF - mF2)
	}
	h[100] = 2 * math.Pi * fc
	var sum float64 = 0
	for i := 0; i <= m; i++ {
		sum += h[i]
	}
	for i := 0; i <= m; i++ {
		h[i] /= sum
	}
	return h
}

func windowSincKernelHp(m int, fc float64) []float64 {
	hp := windowSincKernelLp(m, fc)
	for i := 0; i < len(hp); i++ {
		hp[i] = -hp[i]
	}
	hp[len(hp)/2] += 1
	return hp
}

func windowSincKernelBp(m int, fcL, fcH float64) []float64 {
	lp := windowSincKernelLp(m, fcL)
	hp := windowSincKernelHp(m, fcH)
	bp := make([]float64, m+1)
	for i := 0; i < len(bp); i++ {
		bp[i] = lp[i] + hp[i]
		bp[i] = -bp[i]
	}
	bp[len(bp)/2] += 1
	return bp
}

func fftConvolve(data []float64, filter []float64) []float64 {
	n := len(data)
	m := len(filter)
	dataC := dsputils.ZeroPad(dsputils.ToComplex(data), n+m-1)
	filterC := dsputils.ZeroPad(dsputils.ToComplex(filter), n+m-1)

	result := fft.Convolve(filterC, dataC)
	return ToAbs(result)
}

func testFft() {
	data := []float64{0, 5, 10, 5, 0}
	filter := []float64{5, 5, 5, 5, 5}

	result := fftConvolve(filter, data)
	fmt.Printf("result: \n")
	for i := 0; i < len(result); i++ {
		fmt.Printf("%0.5f ", result[i])
	}
	fmt.Printf("\n")
}

func processWhole() {
	buf, _ := readFile("short.wav")
	drawChart("signal.html", buf)

	kernel := dsputils.ZeroPadF(windowSincKernelHp(200, 2.0/441), 200+len(buf))
	buf = dsputils.ZeroPadF(buf, 200+len(buf))
	filtered := ToReal(fft.Convolve(dsputils.ToComplex(buf), dsputils.ToComplex(kernel)))

	buf = filtered
	kernel = dsputils.ZeroPadF(windowSincKernelHp(200, 7.0/441), 200+len(buf))
	buf = dsputils.ZeroPadF(buf, 200+len(buf))
	filtered = ToReal(fft.Convolve(dsputils.ToComplex(buf), dsputils.ToComplex(kernel)))
	drawChart("filtered.html", filtered)

	cut := filtered[27400:38360]
	drawCut(cut)
}

func drawCut(cut []float64) {
	drawChart("cut.html", cut)

	for i := 0; i < len(cut)/441; i++ {
		fileName := fmt.Sprintf("%d.html", i)
		segment := cut[i*441 : (i+1)*441]
		hann(segment)
		drawChart(fileName, segment)
		spectrum := ToAbs(fft.FFTReal(segment))
		fileName = fmt.Sprintf("s%d.html", i)
		mx := 0
		for j := 0; j < len(spectrum); j++ {
			if spectrum[mx] < spectrum[j] {
				mx = j
			}
		}
		fmt.Printf("%v ", mx)
		drawChart(fileName, spectrum)
	}
	fmt.Printf("\n")

}

func measureIntervals(s []bool) (es []Element) {
	es = make([]Element, 1)
	es[0].s = s[0]
	es[0].d = 1
	for i := 1; i < len(s); i++ {
		if s[i] == es[len(es)-1].s {
			es[len(es)-1].d++
		} else {
			es = append(es, Element{1, s[i]})
		}
	}
	return
}
