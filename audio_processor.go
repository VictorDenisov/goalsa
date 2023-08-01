package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"math/cmplx"
	"os"
	"sort"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/mjibson/go-dsp/dsputils"
	"github.com/mjibson/go-dsp/fft"
)

type Range struct {
	lb, ub int64
}

func processFile(name string, rng *Range, classRng *Range) (sig []float64, res []float64, values []bool, linSpectra []float64, err error) {

	values = make([]bool, 0)
	linSpectra = make([]float64, 0)

	filter := NewBpFilter(200, 7.0/441, 30.0/441, 441)

	file, err := os.Open(name)
	if err != nil {
		return nil, nil, nil, nil, err
	}
	defer file.Close()
	sig = make([]float64, 0)
	res = make([]float64, 0)

	spectra := make([][]float64, 0)
	for pieceNum := int64(0); ; pieceNum++ {
		buf := make([]float64, 441)
		for i := 0; i < 441; i++ {
			var v int16
			err := binary.Read(file, binary.LittleEndian, &v)
			if err != nil {
				goto exit
			}
			buf[i] = float64(v)
		}
		sig = append(sig, buf...)
		buf = filter.FilterBuf(buf)
		res = append(res, buf...)
		hann(buf)
		rawSpectrum := ToAbs(fft.FFTReal(buf))
		spectra = append(spectra, rawSpectrum[0:222])
		linSpectra = append(linSpectra, rawSpectrum...)
		if rng != nil && pieceNum > rng.lb && pieceNum < rng.ub {
			fn := fmt.Sprintf("%d.html", pieceNum)
			drawChart(fn, rawSpectrum)
		}

	}
exit:
	/*
		var sd *KMeansSignalDetector
		if classRng.lb == 0 && classRng.ub == 0 {
			sd = classifySegments(spectra)
		} else {
			sd = classifySegments(spectra[classRng.lb:classRng.ub])
		}
		drawChart("signalMean.html", sd.signal)
		drawChart("noiseMean.html", sd.noise)
		for i := 0; i < len(spectra); i++ {
			values = append(values, sd.isSignal(spectra[i]))
		}
		return sig, res, values, nil
	*/

	significantFrequency, err := calculateSignificantFrequency(spectra)
	fmt.Printf("Significant frequency result: %v, %v\n", significantFrequency, err)

	signals := make([]float64, len(spectra))
	for i := 0; i < len(spectra); i++ {
		signals[i] = spectra[i][significantFrequency]
	}
	//signals = signals[10:len(signals)]
	//sort.Float64s(signals)
	/*
		for i := 0; i < len(signals); i++ {
			fmt.Printf("%0.6f\n", signals[i])
		}
	*/
	signal := make([]float64, len(spectra))
	copy(signal, signals)
	signal = cleanupSignal(signal)
	sd := classifyEMFromSingleFrequency(signal)
	fmt.Printf("EM Classifier: %v\n", sd)
	/*
		var sd *EMSignalDetector
		if classRng == nil || (classRng.lb == 0 && classRng.ub == 0) {
			sd = expectationMaximizationClassifySegments(spectra)
		} else {
			sd = expectationMaximizationClassifySegments(spectra[classRng.lb:classRng.ub])
		}
	*/

	//drawChart("signalMean.html", sd.centroids[0])
	//drawChart("noiseMean.html", sd.centroids[1])

	for i := 0; i < len(signals); i++ {
		//rk := assignPoint(signals[i], sd.m, sd.sigma, sd.pi)
		//fmt.Printf("%v\n", rk)
		values = append(values, sd.isSignal(signals[i]))
	}

	return sig, res, values, linSpectra, nil
}

func cleanupSignal(signal []float64) []float64 {
	sort.Float64s(signal)
	for {
		allMin := signal[0]
		allMax := signal[len(signal)-1]
		middle := (allMin + allMax) / 2

		aboveMiddle := 0
		belowMiddle := 0

		for _, v := range signal {
			if v > middle {
				aboveMiddle++
			} else {
				belowMiddle++
			}
		}
		if aboveMiddle == 1 {
			signal = signal[0 : len(signal)-1]
		} else if belowMiddle == 1 {
			signal = signal[1:len(signal)]
		} else {
			break
		}
	}
	return signal
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

func calculateSignificantFrequency(spectra [][]float64) (int, error) {
	if len(spectra) == 0 {
		return 0, fmt.Errorf("Not enough data to calculate significant frequency")
	}
	n := len(spectra)
	m := len(spectra[0])
	sum := make([]float64, m)
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			sum[j] += spectra[i][j]
		}
	}
	for j := 0; j < m; j++ {
		sum[j] /= float64(n)
	}
	r := 0
	for j := 1; j < m; j++ {
		if sum[j] > sum[r] {
			r = j
		}
	}
	return r, nil
}
