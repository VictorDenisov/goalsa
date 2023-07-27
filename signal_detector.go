package main

import (
	_ "fmt"
	"math"
)

type KMeansSignalDetector struct {
	signal []float64
	noise  []float64
}

func (sd *KMeansSignalDetector) isSignal(sample []float64) bool {
	return distSq(sample, sd.signal) < distSq(sample, sd.noise)
}

type EMSignalDetector struct {
	r         [][]float64
	centroids [][]float64
	sigma     []float64
	pi        []float64
}

func (sd *EMSignalDetector) isSignal(sample []float64) bool {
	return distSq(sample, sd.centroids[0]) < distSq(sample, sd.centroids[1])
}

func classifySegments(segments [][]float64) (sd *KMeansSignalDetector) {

	m := len(segments[0])
	n := len(segments)

	labels := make([]bool, n)

	mx := make([]float64, n)

	for i := 0; i < n; i++ {
		mx[i], _ = segmentMax(segments[i])
	}

	allMin, _ := segmentMin(mx)
	allMax, _ := segmentMax(mx)

	middle := (allMin + allMax) / 2

	for i := 0; i < n; i++ {
		labels[i] = mx[i] > middle
	}

	signalMean := make([]float64, m)
	noiseMean := make([]float64, m)
	for {
		signalMean = make([]float64, m)
		noiseMean = make([]float64, m)

		signalNum := 0
		noiseNum := 0
		for i := 0; i < n; i++ {
			if labels[i] {
				add(signalMean, segments[i])
				signalNum++
			} else {
				add(noiseMean, segments[i])
				noiseNum++
			}
		}

		div(signalMean, float64(signalNum))
		div(noiseMean, float64(noiseNum))

		labelChanged := false
		for i := 0; i < n; i++ {
			newLabel := distSq(signalMean, segments[i]) < distSq(noiseMean, segments[i])
			labelChanged = labelChanged || (labels[i] != newLabel)
			labels[i] = newLabel
		}
		if !labelChanged {
			break
		}
	}

	sd = &KMeansSignalDetector{signalMean, noiseMean}
	return sd
}

func expectationMaximizationClassifySegments(segments [][]float64) (sd *EMSignalDetector) {

	sq2pi := math.Sqrt(2 * math.Pi)
	m := len(segments[0])
	n := len(segments)

	r := make([][]float64, n)
	for i := 0; i < n; i++ {
		r[i] = make([]float64, 2)
	}

	mx := make([]float64, n) // Maximum in every segment

	for i := 0; i < n; i++ {
		mx[i], _ = segmentMax(segments[i])
	}

	_, minId := segmentMin(mx)
	_, maxId := segmentMax(mx)

	const nCentroids = 2
	centroids := make([][]float64, nCentroids)
	for i := 0; i < nCentroids; i++ {
		centroids[i] = make([]float64, m)
	}
	copy(centroids[0], segments[minId])
	copy(centroids[1], segments[maxId])
	drawChart("signalMean0.html", centroids[0])
	drawChart("noiseMean0.html", centroids[1])
	sigma := make([]float64, nCentroids)
	for i := 0; i < nCentroids; i++ {
		sigma[i] = 100
	}
	pi := make([]float64, nCentroids)
	for i := 0; i < nCentroids; i++ {
		pi[i] = 1
	}
	for i := 0; i < n; i++ {
		//sumR := float64(0)
		//fmt.Printf("dist: ")
		for k := 0; k < nCentroids; k++ {
			//v := pi[k] * math.Exp(-dist(centroids[k], segments[i])/sigma[k]/sigma[k])
			//v = float64(m) //sq2pi * sigma[k]
			//fmt.Printf("%f ", v)
			r[i][k] = pi[k] * math.Exp(-dist(centroids[k], segments[i])/sigma[k]/sigma[k]) / math.Pow(sq2pi*sigma[k], float64(m))
			//sumR = r[i][k]
		}
		//fmt.Printf("\n")

		for k := 0; k < nCentroids; k++ {
			//r[i][k] /= sumR
			//fmt.Printf("%f ", r[i][k])
		}
		//fmt.Printf("\n")
	}
	R := make([]float64, nCentroids)
	totalR := float64(0)
	for k := 0; k < nCentroids; k++ {
		s := make([]float64, m)
		for j := 0; j < n; j++ {
			s = plus(s, mult(r[j][k], segments[j]))
			R[k] += r[j][k]
		}
		centroids[k] = divN(s, R[k])
		sigma[k] = 0
		for j := 0; j < n; j++ {
			sigma[k] += r[j][k] * dist(segments[k], centroids[k])
		}
		sigma[k] /= float64(m) * R[k]
		sigma[k] = math.Sqrt(sigma[k])
		totalR += R[k]
	}
	for k := 0; k < nCentroids; k++ {
		pi[k] = R[k] / totalR
	}
	return &EMSignalDetector{r, centroids, sigma, pi}
}

type SingleFrequencyDetector struct {
	signal float64
	noise  float64
}

func (sd *SingleFrequencyDetector) isSignal(sample float64) bool {
	return math.Abs(sample-sd.signal) < math.Abs(sample-sd.noise)
}

func classifyFromSingleFrequency(signals []float64) (sd *SingleFrequencyDetector) {
	n := len(signals)

	labels := make([]bool, n)

	allMin, _ := segmentMin(signals)
	allMax, _ := segmentMax(signals)

	middle := (allMin + allMax) / 2

	for i := 0; i < n; i++ {
		labels[i] = signals[i] > middle
	}

	var signalMean float64
	var noiseMean float64
	for {
		signalMean = 0
		noiseMean = 0

		signalNum := 0
		noiseNum := 0
		for i := 0; i < n; i++ {
			if labels[i] {
				signalMean += signals[i]
				signalNum++
			} else {
				noiseMean += signals[i]
				noiseNum++
			}
		}

		signalMean /= float64(signalNum)
		noiseMean /= float64(noiseNum)

		labelChanged := false
		for i := 0; i < n; i++ {
			newLabel := math.Abs(signalMean-signals[i]) < math.Abs(noiseMean-signals[i])
			labelChanged = labelChanged || (labels[i] != newLabel)
			labels[i] = newLabel
		}
		if !labelChanged {
			break
		}
	}

	sd = &SingleFrequencyDetector{signalMean, noiseMean}
	return sd
}

func plus(a, b []float64) (r []float64) {
	r = make([]float64, len(b))
	for i := 0; i < len(b); i++ {
		r[i] = a[i] + b[i]
	}
	return
}

func divN(a []float64, b float64) (r []float64) {
	r = make([]float64, len(a))
	for i := 0; i < len(a); i++ {
		r[i] = a[i] / b
	}
	return
}

func mult(a float64, b []float64) (r []float64) {
	r = make([]float64, len(b))
	for i := 0; i < len(b); i++ {
		r[i] = b[i] * a
	}
	return
}

func dist(a, b []float64) (r float64) {
	for i := 0; i < len(a); i++ {
		r += (a[i] - b[i]) * (a[i] - b[i])
	}
	return math.Sqrt(r)
}

func distSq(x, y []float64) float64 {
	res := float64(0)
	for i := 0; i < len(x); i++ {
		res += (y[i] - x[i]) * (y[i] - x[i])
	}
	return res
}

func add(x, y []float64) {
	n := len(x)
	for i := 0; i < n; i++ {
		x[i] += y[i]
	}
}

func div(x []float64, v float64) {
	n := len(x)
	for i := 0; i < n; i++ {
		x[i] /= v
	}
}

func segmentMin(seg []float64) (mx float64, id int) {
	mx = seg[0]
	id = 0
	for i := 1; i < len(seg); i++ {
		if seg[i] < mx {
			mx = seg[i]
			id = i
		}
	}
	return
}

func segmentMax(seg []float64) (mx float64, id int) {
	mx = seg[0]
	id = 0
	for i := 1; i < len(seg); i++ {
		if seg[i] > mx {
			mx = seg[i]
		}
	}
	return mx, id
}
